package models

type ProfileData struct {
	Hash       		string `json:"hash"`
	Email       	string `json:"email"`
	Uid         	string `json:"uid"`
	First_name  	string `json:"first_name"`
	Last_name   	string `json:"last_name"`
	Photo       	string `json:"photo"`
	Age       		string `json:"age"`
	City        	string `json:"city"`
	Country     	string `json:"country"`
	Status 			string `json:"status"` 	// - src поля Status в профиле (иногда необходимо для доп.фильтрации)
	Raw	       		[]Data `json:"raw"`	// объект пользователя (нужен при сборки проекта для данного юзера при добавлении прав на базу)
	Tables      	[]Data `json:"tables"`
	Roles       	[]Data `json:"roles"`
	Homepage		string `json:"homepage"`
	Maket			string `json:"maket"`
	UpdateFlag 		bool `json:"update_flag"`
	UpdateData 		[]Data `json:"update_data"`
	CurrentRole 	Data `json:"current_role"`
	CurrentProfile 	Data `json:" "`
	Navigator   	[]*Items `json:"navigator"`
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


