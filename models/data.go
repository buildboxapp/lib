package models

type Data struct {
	Uid        		string               `json:"uid"`
	Id         		string               `json:"id"`
	Source     		string               `json:"source"`
	Parent     		string               `json:"parent"`
	Type       		string               `json:"type"`
	Title      		string               `json:"title"`
	Rev        		string               `json:"rev"`
	Сopies			string 				 `json:"copies"`
	Attributes 		map[string]Attribute `json:"attributes"`
}

type Attribute struct {
	Value  string `json:"value"`
	Src    string `json:"src"`
	Tpls   string `json:"tpls"`
	Status string `json:"status"`
	Rev    string `json:"rev"`
	Editor string `json:"editor"`
}

type Response struct {
	Data   	interface{} 	`json:"data"`
	Status 	RestStatus    	`json:"status"`
	//Metrics Metrics 		`json:"metrics"`
}

type ResponseData struct {
	Data      []Data        `json:"data"`
	Res   	  interface{} 	`json:"res"`
	Status    RestStatus    `json:"status"`
	Metrics   Metrics 		`json:"metrics"`
}

type Metrics struct {
	ResultSize     	int `json:"result_size"`
	ResultCount     int `json:"result_count"`
	ResultOffset    int `json:"result_offset"`
	ResultLimit     int `json:"result_limit"`
	ResultPage 		int `json:"result_page"`
	TimeExecution   string `json:"time_execution"`
	TimeQuery   	string `json:"time_query"`

	PageLast		int `json:"page_last"`
	PageCurrent		int `json:"page_current"`
	PageList		[]int `json:"page_list"`
	PageFrom		int `json:"page_from"`
	PageTo			int `json:"page_to"`
}
