package models

type ProfileData struct {
	Hash       		string
	Email       	string
	Uid         	string
	First_name  	string
	Last_name   	string
	Photo       	string
	Age       		string
	City        	string
	Country     	string
	Status 			string 	// - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw	       		[]Data	// объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables      	[]Data
	Roles       	[]Data
	Homepage		string
	Maket			string
	UpdateFlag 		bool
	UpdateData 		[]Data
	CurrentRole 	Data
	CurrentProfile 	Data
	Navigator   	[]*Items
}


type Items struct {
	Title  			string   	`json:"title"`
	ExtentedLink 	string 		`json:"extentedLink"`
	Uid    			string   	`json:"uid"`
	Source 			string   	`json:"source"`
	Icon   			string   	`json:"icon"`
	Leader 			string   	`json:"leader"`
	Order  			string   	`json:"order"`
	Type   			string   	`json:"type"`
	Preview			string   	`json:"preview"`
	Url    			string   	`json:"url"`
	Sub    			[]string 	`json:"sub"`
	Incl   			[]*Items 	`json:"incl"`
	Class 			string 		`json:"class"`
}


