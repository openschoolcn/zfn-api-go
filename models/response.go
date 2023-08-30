package models

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type LoginKaptcha struct {
	Sid        string            `json:"sid"`
	CsrfToken  string            `json:"csrf_token"`
	Cookies    map[string]string `json:"cookies"`
	Modulus    string            `json:"modulus"`
	Exponent   string            `json:"exponent"`
	KaptchaPic string            `json:"kaptcha_pic"`
	Timestamp  string            `json:"timestamp"`
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

type GradeCourse struct {
	CourseId        string `json:"course_id"`
	Title           string `json:"title"`
	Teacher         string `json:"teacher"`
	ClassName       string `json:"class_name"`
	Credit          string `json:"credit"`
	Category        string `json:"category"`
	Nature          string `json:"nature"`
	Grade           string `json:"grade"`
	GradePoint      string `json:"grade_point"`
	GradeNature     string `json:"grade_nature"`
	TeachingCollege string `json:"teaching_college"`
	Mark            string `json:"mark"`
}

type GradeInfo struct {
	Year    int           `json:"year"`
	Term    int           `json:"term"`
	Count   int           `json:"count"`
	Courses []GradeCourse `json:"courses"`
}
