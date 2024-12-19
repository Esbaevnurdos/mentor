package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Esbaevnurdos/models"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StudentController struct {
	Client *mongo.Client
}

func (sc *StudentController) CreateStudent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Address    string `json:"address"`
		ClassName  string `json:"class_name"`
		GradeLevel string `json:"grade_level"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	studentCollection := sc.Client.Database("school").Collection("students")
	classCollection := sc.Client.Database("school").Collection("classes")
	gradeLevelCollection := sc.Client.Database("school").Collection("grade_levels")

	// Create new student object with generated ID
	studentID := primitive.NewObjectID()
	student := models.Student{
		ID:         studentID,
		FirstName:  input.FirstName,
		LastName:   input.LastName,
		Address:    input.Address,
		GradeLevel: input.GradeLevel,
		ClassName:  input.ClassName,
	}

	// Insert student into students collection
	_, err = studentCollection.InsertOne(ctx, student)
	if err != nil {
		http.Error(w, "Failed to create student", http.StatusInternalServerError)
		return
	}

	// Create corresponding grade level and insert into grade_levels collection
	gradeLevel := models.GradeLevel{
		StudentID: studentID.Hex(),
		Level:     input.GradeLevel,
	}

	_, err = gradeLevelCollection.InsertOne(ctx, gradeLevel)
	if err != nil {
		http.Error(w, "Failed to create grade level", http.StatusInternalServerError)
		return
	}

	// Create corresponding class and insert into classes collection
	class := models.Class{
		StudentID: studentID.Hex(),
		ClassName: input.ClassName,
		GradeLevel: input.GradeLevel,
	}

	_, err = classCollection.InsertOne(ctx, class)
	if err != nil {
		http.Error(w, "Failed to create class", http.StatusInternalServerError)
		return
	}

	// Return success response with the created student data
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(student)
}

func (sc *StudentController) UpdateStudent(w http.ResponseWriter, r *http.Request) {
	// Extract student ID from the URL
	vars := mux.Vars(r)
	studentID := vars["student_id"]

	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	// Convert studentID to MongoDB ObjectID (if stored as ObjectID)
	objectID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		http.Error(w, "Invalid student_id format", http.StatusBadRequest)
		return
	}

	// Parse the input data
	var input struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		Address    string `json:"address"`
		ClassName  string `json:"class_name"`
		GradeLevel string `json:"grade_level"`
	}

	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Access collections
	studentCollection := sc.Client.Database("school").Collection("students")
	classCollection := sc.Client.Database("school").Collection("classes")
	gradeLevelCollection := sc.Client.Database("school").Collection("grade_levels")

	// Update student fields in the `students` collection
	updateStudent := bson.M{"$set": bson.M{
		"first_name":  input.FirstName,
		"last_name":   input.LastName,
		"address":     input.Address,
		"class_name":  input.ClassName,
		"grade_level": input.GradeLevel,
	}}

	// Perform the update in the student collection
	result, err := studentCollection.UpdateOne(ctx, bson.M{"_id": objectID}, updateStudent)
	if err != nil || result.MatchedCount == 0 {
		http.Error(w, "Student not found or failed to update", http.StatusNotFound)
		return
	}

	// Update the `classes` collection
	_, err = classCollection.UpdateOne(ctx, bson.M{"student_id": studentID}, bson.M{"$set": bson.M{
		"class_name":  input.ClassName,
		"grade_level": input.GradeLevel,
	}})
	if err != nil {
		http.Error(w, "Failed to update class", http.StatusInternalServerError)
		return
	}

	// Update the `grade_levels` collection
	_, err = gradeLevelCollection.UpdateOne(ctx, bson.M{"student_id": studentID}, bson.M{"$set": bson.M{
		"level": input.GradeLevel,
	}})
	if err != nil {
		http.Error(w, "Failed to update grade level", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student updated successfully"})
}


func (sc *StudentController) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	// Extract student ID from the URL
	vars := mux.Vars(r)
	studentID := vars["student_id"]

	// Ensure studentID is provided
	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	// Convert studentID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		http.Error(w, "Invalid student_id format", http.StatusBadRequest)
		return
	}

	// Set up context with timeout for MongoDB operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Access the student collection in MongoDB
	studentCollection := sc.Client.Database("school").Collection("students")

	// Find the student by ObjectID
	var student models.Student
	err = studentCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&student)
	if err != nil {
		// If student is not found
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Student not found", http.StatusNotFound)
		} else {
			// Handle other errors
			http.Error(w, "Error retrieving student", http.StatusInternalServerError)
		}
		return
	}

	// Return the student as a JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(student); err != nil {
		http.Error(w, "Failed to encode student: "+err.Error(), http.StatusInternalServerError)
	}
}

func (sc *StudentController) DeleteStudent(w http.ResponseWriter, r *http.Request) {
	// Extract student ID from the URL
	vars := mux.Vars(r)
	studentID := vars["student_id"]

	// Ensure studentID is provided
	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	// Convert studentID to MongoDB ObjectID
	objectID, err := primitive.ObjectIDFromHex(studentID)
	if err != nil {
		http.Error(w, "Invalid student_id format", http.StatusBadRequest)
		return
	}

	// Set up context with timeout for MongoDB operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Access the student collection in MongoDB
	studentCollection := sc.Client.Database("school").Collection("students")
	classCollection := sc.Client.Database("school").Collection("classes")
	gradeLevelCollection := sc.Client.Database("school").Collection("grade_levels")

	// Delete the student from the students collection
	_, err = studentCollection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		http.Error(w, "Failed to delete student", http.StatusInternalServerError)
		return
	}

	// Delete the grade level associated with the student
	_, err = gradeLevelCollection.DeleteOne(ctx, bson.M{"student_id": objectID})
	if err != nil {
		http.Error(w, "Failed to delete grade level", http.StatusInternalServerError)
		return
	}

	// Delete the class associated with the student
	_, err = classCollection.DeleteOne(ctx, bson.M{"student_id": objectID})
	if err != nil {
		http.Error(w, "Failed to delete class", http.StatusInternalServerError)
		return
	}

	// Send a success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student deleted successfully"})
}


func (sc *StudentController) GetClassByStudentID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID := vars["student_id"]
	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	classCollection := sc.Client.Database("school").Collection("classes")

	var class models.Class
	err := classCollection.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&class)
	if err != nil {
		http.Error(w, "class not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(class)
}

func (sc *StudentController) GetGradeLevelByStudentID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	studentID := vars["student_id"]
	if studentID == "" {
		http.Error(w, "student_id is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	gradeLevelCollection := sc.Client.Database("school").Collection("grade_levels")

	var gradeLevel models.GradeLevel
	err := gradeLevelCollection.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&gradeLevel)
	if err != nil {
		http.Error(w, "grade level not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(gradeLevel)
}


func (sc *StudentController) SearchStudents(w http.ResponseWriter, r *http.Request) {
	// Extract query parameters
	class := r.URL.Query().Get("class_name")
	gradeLevel := r.URL.Query().Get("grade_level")
	firstName := r.URL.Query().Get("first_name")
	lastName := r.URL.Query().Get("last_name")

	// Initialize an empty filter
	filter := bson.M{}

	// Build the filter based on query parameters
	if class != "" {
		filter["class_name"] = bson.M{"$regex": class, "$options": "i"} // Case-insensitive search
	}
	if gradeLevel != "" {
		filter["grade_level"] = bson.M{"$regex": gradeLevel, "$options": "i"} // Case-insensitive search
	}
	if firstName != "" {
		filter["first_name"] = bson.M{"$regex": firstName, "$options": "i"} // Case-insensitive search
	}
	if lastName != "" {
		filter["last_name"] = bson.M{"$regex": lastName, "$options": "i"} // Case-insensitive search
	}

	// Log the filter for debugging purposes
	fmt.Println("Search filter:", filter)

	// Set up a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Access the student collection
	studentCollection := sc.Client.Database("school").Collection("students")

	// Perform the search with the filter
	cursor, err := studentCollection.Find(ctx, filter)
	if err != nil {
		http.Error(w, "Failed to search students: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Initialize a slice to store results
	var students []models.Student

	// Iterate through the results
	for cursor.Next(ctx) {
		var student models.Student
		if err := cursor.Decode(&student); err != nil {
			http.Error(w, "Failed to decode student: "+err.Error(), http.StatusInternalServerError)
			return
		}
		students = append(students, student)
	}

	// Check for errors in the cursor iteration
	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// If no students were found, return a 404 response
	if len(students) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "No students found"})
		return
	}

	// Return the students as a JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(students); err != nil {
		http.Error(w, "Failed to encode students: "+err.Error(), http.StatusInternalServerError)
	}
}

