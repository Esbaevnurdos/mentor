package routes

import (
	"github.com/Esbaevnurdos/controllers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)
func RegisterStudentRoutes(router *mux.Router, client *mongo.Client) {
	studentController := controllers.StudentController{Client: client}

	router.HandleFunc("/students", studentController.CreateStudent).Methods("POST")
	router.HandleFunc("/students/{student_id}", studentController.GetStudentByID).Methods("GET")
	router.HandleFunc("/students/{student_id}", studentController.UpdateStudent).Methods("PUT")
	router.HandleFunc("/students/{student_id}", studentController.DeleteStudent).Methods("DELETE")
	router.HandleFunc("/class/{student_id}", studentController.GetClassByStudentID).Methods("GET")
	router.HandleFunc("/grade-level/{student_id}", studentController.GetGradeLevelByStudentID).Methods("GET")

	// Updated search endpoint
router.HandleFunc("/search-students", studentController.SearchStudents).Methods("GET")
}
