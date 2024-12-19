package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Student struct {
	ID         primitive.ObjectID `bson:"_id"`
	FirstName  string             `bson:"first_name"`
	LastName   string             `bson:"last_name"`
	Address    string             `bson:"address"`
	GradeLevel string             `bson:"grade_level"` 
	ClassName  string             `bson:"class_name"`   
}

type Class struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	StudentID string             `bson:"student_id"`  
	ClassName string             `bson:"class_name"`
	GradeLevel string            `bson:"grade_level"`
}

type GradeLevel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	StudentID string             `bson:"student_id"`  
	Level     string             `bson:"level"`      
}
