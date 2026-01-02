package models

import "gorm.io/gorm"

type Classroom struct {
	gorm.Model
	Code      string            `json:"code" gorm:"unique;not null"`
	Name      string            `json:"name"`
	TeacherID uint              `json:"teacher_id"`
	Teacher   User              `json:"teacher" gorm:"foreignKey:TeacherID"`
	Members   []ClassroomMember `json:"members" gorm:"foreignKey:ClassroomID"`
}

type ClassroomMember struct {
	gorm.Model
	ClassroomID uint      `json:"classroom_id"`
	Classroom   Classroom `json:"classroom" gorm:"foreignKey:ClassroomID"`
	StudentID   uint      `json:"student_id"`
	Student     User      `json:"student" gorm:"foreignKey:StudentID"`
}

type Assignment struct {
	gorm.Model
	ClassroomID uint      `json:"classroom_id"`
	Classroom   Classroom `json:"-" gorm:"foreignKey:ClassroomID"`
	QuizID      uint      `json:"quiz_id"`
	Quiz        Quiz      `json:"quiz" gorm:"foreignKey:QuizID"`
	Deadline    string    `json:"deadline"` // Format: YYYY-MM-DD HH:mm:ss
}
