package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

func generateClassCode() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "ABCDEF"
	}
	return hex.EncodeToString(bytes)
}

func stringToUint(s string) uint {
	val, _ := strconv.Atoi(s)
	return uint(val)
}

// CreateClassroom (Teacher only)
// CreateClassroom (Teacher or Admin)
func CreateClassroom(c *fiber.Ctx) error {
	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	// Determine if User or Admin
	// Middleware usually sets "user_id" and "role" (for admin) or just "user_id" (for user)?
	// Let's check how LoginAdmin sets context. Usually it sets "user" or "role".
	// Assuming "role" exists for Admin.

	role := c.Locals("role") // "supervisor" or "admin" from JWT?
	userID := uint(c.Locals("user_id").(float64))

	classroom := models.Classroom{
		Code: generateClassCode(),
		Name: input.Name,
	}

	if role != nil {
		// It is an Admin
		classroom.AdminID = &userID
	} else {
		// It is a User (Teacher)
		classroom.TeacherID = &userID
	}

	if err := config.DB.Create(&classroom).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create classroom", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Classroom created", classroom)
}

// JoinClassroom (Student)
func JoinClassroom(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	var input struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var classroom models.Classroom
	if err := config.DB.Where("code = ?", input.Code).First(&classroom).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Classroom not found", nil)
	}

	// Check if already a member
	var member models.ClassroomMember
	if err := config.DB.Where("classroom_id = ? AND student_id = ?", classroom.ID, userID).First(&member).Error; err == nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Already joined this class", nil)
	}

	newMember := models.ClassroomMember{
		ClassroomID: classroom.ID,
		StudentID:   userID,
	}

	if err := config.DB.Create(&newMember).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to join", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Joined classroom successfully", classroom)
}

// GetMyClassrooms (Student & Teacher)
func GetMyClassrooms(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	// If Teacher: Get classes they created
	var teaching []models.Classroom
	config.DB.Where("teacher_id = ?", userID).Find(&teaching)

	// If Student: Get classes they joined
	var joining []models.Classroom
	config.DB.Joins("JOIN classroom_members ON classroom_members.classroom_id = classrooms.id").
		Where("classroom_members.student_id = ?", userID).
		Find(&joining)

	return utils.SuccessResponse(c, fiber.StatusOK, "Classrooms retrieved", fiber.Map{
		"teaching": teaching,
		"joining":  joining,
	})
}

// CreateAssignment (Teacher only)
func CreateAssignment(c *fiber.Ctx) error {
	classroomID := c.Params("id")
	var input struct {
		QuizID   uint   `json:"quiz_id"`
		Deadline string `json:"deadline"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	assignment := models.Assignment{
		ClassroomID: stringToUint(classroomID),
		QuizID:      input.QuizID,
		Deadline:    input.Deadline,
	}

	if err := config.DB.Create(&assignment).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to create assignment", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Assignment created", assignment)
}

// GetClassroomDetails (Members only)
func GetClassroomDetails(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := c.Locals("user_id").(float64)

	var classroom models.Classroom
	if err := config.DB.Preload("Teacher").Preload("Members.Student").First(&classroom, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Classroom not found", nil)
	}

	var assignments []models.Assignment
	config.DB.Preload("Quiz").Where("classroom_id = ?", id).Find(&assignments)

	// Check my submissions for these assignments
	mySubmissions := make(map[uint]models.History)
	if len(assignments) > 0 {
		var assignmentIDs []uint
		for _, a := range assignments {
			assignmentIDs = append(assignmentIDs, a.ID)
		}
		var histories []models.History
		config.DB.Where("user_id = ? AND assignment_id IN ?", userID, assignmentIDs).Find(&histories)
		for _, h := range histories {
			if h.AssignmentID != nil {
				mySubmissions[*h.AssignmentID] = h
			}
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Classroom details", fiber.Map{
		"classroom":      classroom,
		"assignments":    assignments,
		"my_submissions": mySubmissions,
	})
}

// GetAssignmentSubmissions (Admin)
func GetAssignmentSubmissions(c *fiber.Ctx) error {
	assignmentID := c.Params("id")
	var histories []models.History
	// preload User to get names
	if err := config.DB.Preload("User").Where("assignment_id = ?", assignmentID).Order("score desc").Find(&histories).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch submissions", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Submissions retrieved", histories)
}

// GetAllClassrooms (Admin only)
func GetAllClassrooms(c *fiber.Ctx) error {
	var classrooms []models.Classroom
	if err := config.DB.Preload("Teacher").Preload("Admin").Preload("Members").Find(&classrooms).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch classrooms", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "All classrooms retrieved", classrooms)
}

// AddClassroomMember (Admin/Teacher manually adds a student)
func AddClassroomMember(c *fiber.Ctx) error {
	var input struct {
		ClassroomID uint   `json:"classroom_id"`
		StudentID   uint   `json:"student_id"`
		Username    string `json:"username"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	// Resolve StudentID from Username if provided
	if input.Username != "" {
		var user models.User
		if err := config.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found with that username", nil)
		}
		input.StudentID = user.ID
	}

	if input.StudentID == 0 {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Student ID or Username is required", nil)
	}

	var classroom models.Classroom
	if err := config.DB.First(&classroom, input.ClassroomID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Classroom not found", nil)
	}

	// Check if already a member
	var member models.ClassroomMember
	if err := config.DB.Where("classroom_id = ? AND student_id = ?", input.ClassroomID, input.StudentID).First(&member).Error; err == nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "User already in this class", nil)
	}

	newMember := models.ClassroomMember{
		ClassroomID: input.ClassroomID,
		StudentID:   input.StudentID,
	}

	if err := config.DB.Create(&newMember).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to add member", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusCreated, "Member added successfully", newMember)
}

// RemoveClassroomMember (Teacher/Admin kicks student)
func RemoveClassroomMember(c *fiber.Ctx) error {
	classroomID := c.Params("id")
	studentID := c.Params("studentId")

	if err := config.DB.Where("classroom_id = ? AND student_id = ?", classroomID, studentID).Delete(&models.ClassroomMember{}).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to remove member", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Member removed", nil)
}

// DeleteAssignment (Teacher/Admin)
func DeleteAssignment(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Assignment{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete assignment", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Assignment deleted", nil)
}

func AssignClassroomTeacher(c *fiber.Ctx) error {
	id := c.Params("id")
	var input struct {
		Username string `json:"username"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var user models.User
	if err := config.DB.Where("username = ?", input.Username).First(&user).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Teacher username not found", nil)
	}

	var classroom models.Classroom
	if err := config.DB.First(&classroom, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Classroom not found", nil)
	}

	classroom.TeacherID = &user.ID
	if err := config.DB.Save(&classroom).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to assign teacher", err.Error())
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Teacher assigned successfully", classroom)
}

func DeleteClassroom(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := config.DB.Delete(&models.Classroom{}, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed delete classroom", err.Error())
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Classroom deleted", nil)
}
