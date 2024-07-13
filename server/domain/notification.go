package domain

type Notification struct {
	Time    string `json:"time" bson:"time"`
	Message string `json:"message" bson:"message"`
	Email   string `json:"email" bson:"email"`
}
