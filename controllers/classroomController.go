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
func CreateClassroom(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))

	var input struct {
		Name string `json:"name"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	classroom := models.Classroom{
		Code:      generateClassCode(),
		Name:      input.Name,
		TeacherID: userID,
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

	var classroom models.Classroom
	if err := config.DB.Preload("Teacher").Preload("Members.Student").First(&classroom, id).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Classroom not found", nil)
	}

	var assignments []models.Assignment
	config.DB.Preload("Quiz").Where("classroom_id = ?", id).Find(&assignments)

	return utils.SuccessResponse(c, fiber.StatusOK, "Classroom details", fiber.Map{
		"classroom":   classroom,
		"assignments": assignments,
	})
}
