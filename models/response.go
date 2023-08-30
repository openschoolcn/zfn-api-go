package models

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type StudentInfo struct {
	Sid             string `json:"sid"`
	Name            string `json:"name"`
	Domicile        string `json:"domicile"`
	PhoneNumber     string `json:"phone_number"`
	Email           string `json:"email"`
	PoliticalStatus string `json:"political_status"`
	Nationality     string `json:"nationality"`
	CollegeName     string `json:"college_name"`
	MajorName       string `json:"major_name"`
	ClassName       string `json:"class_name"`
}
