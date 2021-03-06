package controller

import (
	"html/template"
	"net/http"
	"encoding/json"
	"strconv"
)

//The page handle
type Handle struct {
	//User processor
	user User
	//language configuration processor
	lang Language
	//Database processor
	db *Database
}

func (this *Handle) Init(db *Database) {
	//Save the database processor
	this.db = db
	//Initialize the user processor
	this.user.Init(db, 3600)
}

/////////////////////////////////////
//This part is a generic module
/////////////////////////////////////

//Get the template file path
func (this *Handle) GetTempSrc(name string) string {
	return "template" + GetPathSep() + name
}

//Output text directly to the browser
func (this *Handle) PostText(w http.ResponseWriter, r *http.Request, content string) {
	var contentByte []byte = []byte(content)
	_, err := w.Write(contentByte)
	if err != nil {
		log.NewLog("You can not directly output string data.", err)
		return
	}
}

//Jump to URL
func (this *Handle) ToURL(w http.ResponseWriter, r *http.Request, urlName string) {
	http.Redirect(w, r, urlName, http.StatusFound)
}

//Output template
func (this *Handle) ShowTemplate(w http.ResponseWriter, r *http.Request, templateFileName string, data interface{}) {
	t, err := template.ParseFiles(this.GetTempSrc(templateFileName),this.GetTempSrc("page-header.html"),this.GetTempSrc("page-menu.html"),this.GetTempSrc("page-footer.html"),this.GetTempSrc("page-menu-nologin.html"))
	if err != nil {
		log.NewLog("The template does not output properly,template file name : "+templateFileName, err)
		return
	}
	if data == nil{
		data = map[string]string{
			"debug" : configData["debug"].(string),
		}
	}
	t.Execute(w, data)
}

//Output the prompt page
func (this *Handle) showTip(w http.ResponseWriter, r *http.Request, title string, contentTitle string, content string, gotoURL string) {
	data := map[string]string{
		"title":        title,
		"contentTitle": contentTitle,
		"content":      content,
		"gotoURL":      gotoURL,
		"debug" : configData["debug"].(string),
	}
	this.ShowTemplate(w, r, "tip.html", data)
}

//Common JSON processing
// w http.ResponseWriter
// r *http.Request
// data interface{} -The data to be sent
// b bool - Whether to run successfully
func (this *Handle) postJSONData(w http.ResponseWriter, r *http.Request,data interface{},b bool) {
	res := make(map[string]interface{})
	res["result"] = b
	res["data"] = data
	res["login"] = this.user.CheckLogin(w, r)
	resJson,err := json.Marshal(res)
	if err != nil{
		log.NewLog("",err)
		this.PostText(w, r, "{'result':false,'data':''}")
	}else{
		resJsonC := string(resJson)
		this.PostText(w, r, resJsonC)
	}
}

//Check that you are logged in
func (this *Handle) CheckLogin(w http.ResponseWriter, r *http.Request) bool {
	if this.user.CheckLogin(w, r) == false {
		log.NewLog("User has not logged in, but visited the home page.", nil)
		this.ToURL(w, r, "/login")
		return false
	}
	return true
}

//Check the post data
func (this *Handle) CheckURLPost(r *http.Request) bool {
	err = r.ParseForm()
	if err != nil {
		log.NewLog("Failed to get get / post data.", err)
		return false
	}
	return true
}

//Update the language data
func (this *Handle) UpdateLanguage() {
	//Initialize the language configuration processor
	this.lang.Init(configData["language"].(string))
	//Set the collector language
	coll.lang = &this.lang
}

/////////////////////////////////////
//This section is the page
/////////////////////////////////////

//404 error handling
func (this *Handle) page404(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		if this.CheckLogin(w, r) == false {
			return
		} else {
			this.ToURL(w, r, "/center")
		}
	} else {
		log.NewLog("The page can not be found,url path : "+r.URL.Path, nil)
		this.ShowTemplate(w, r, "404.html", nil)
	}
}

//Resolve the login page
func (this *Handle) pageLogin(w http.ResponseWriter, r *http.Request) {
	if this.user.CheckLogin(w, r) == true {
		this.ToURL(w, r, "/center")
		return
	} else {
		this.ShowTemplate(w, r, "login.html", nil)
		return
	}
}

//Get the site icon file
func (this *Handle) pageFavicon(w http.ResponseWriter, r *http.Request) {
	this.ToURL(w, r, "/assets/favicon.ico")
}

//Output the set page
func (this *Handle) pageSet(w http.ResponseWriter, r *http.Request) {
	if this.CheckLogin(w, r) == false {
		return
	}
	this.UpdateLanguage()
	this.ShowTemplate(w, r, "set.html", nil)
}

//Output the center page
func (this *Handle) pageCenter(w http.ResponseWriter, r *http.Request) {
	if this.CheckLogin(w, r) == false {
		return
	}
	this.UpdateLanguage()
	this.ShowTemplate(w, r, "center.html", nil)
}

/////////////////////////////////////
//This section is the feedback page
/////////////////////////////////////

//Submit data Try to log in
func (this *Handle) actionLogin(w http.ResponseWriter, r *http.Request) {
	postUser := r.FormValue("email")
	postPasswd := r.FormValue("password")
	b := this.user.LoginIn(w, r, postUser, postPasswd)
	if b == false {
		this.ToURL(w, r, "/login")
		return
	} else {
		this.ToURL(w, r, "/center")
	}
}

//sign out
func (this *Handle) actionLogout(w http.ResponseWriter, r *http.Request) {
	if this.user.CheckLogin(w, r) == false {
		this.ToURL(w, r, "/login")
		return
	}
	b := this.user.Logout(w,r)
	if b == false{
		//...
	}
	this.showTip(w, r, this.lang.Get("handle-logout-title"), this.lang.Get("handle-logout-contentTitle"), this.lang.Get("handle-logout-content"), "/login")
}

//Resolution settings page
func (this *Handle) actionSet(w http.ResponseWriter, r *http.Request) {
	//If not, jump
	if this.CheckLogin(w, r) == false {
		return
	}
	//Make sure that post / get is fine
	b := this.CheckURLPost(r)
	if b == false {
		return
	}
	//Gets the submit action type
	postAction := r.FormValue("action")
	switch postAction {
	case "coll":
		postName := r.FormValue("name")
		if postName == ""{
			return
		}
		if postName == "run-all" {
			coll.Run("")
		}else{
			coll.Run(postName)
		}
		this.postJSONData(w,r,"",true)
		break
	case "get-status":
		data,b := coll.GetStatus()
		this.postJSONData(w,r,data,b)
		break
	case "clear":
		postName := r.FormValue("name")
		if postName == ""{
			return
		}
		this.postJSONData(w,r,coll.ClearColl(postName),true)
		break
	case "clear-log":
		postName := r.FormValue("name")
		if postName == ""{
			return
		}
		this.postJSONData(w,r,coll.ClearLog(postName),true)
		break
	case "close":
		postName := r.FormValue("name")
		if postName == ""{
			return
		}
		this.postJSONData(w,r,coll.ChangeStatus(postName,false),true)
		break
	default:
		this.postJSONData(w,r,"",false)
		return
		break
	}
}

//Feedback center action
func (this *Handle) actionCenter(w http.ResponseWriter, r *http.Request) {
	if this.CheckLogin(w, r) == false {
		return
	}
	this.UpdateLanguage()
}

//Feedback center view content action
func (this *Handle) actionViewList(w http.ResponseWriter, r *http.Request) {
	if this.CheckLogin(w, r) == false {
		return
	}
	//Make sure that post / get is fine
	b := this.CheckURLPost(r)
	if b == false {
		return
	}
	this.UpdateLanguage()
	//get post
	// need : coll string \ parent int64 \ star int \ page int \ max int \ sort int \ desc bool
	postCollName := r.FormValue("coll")
	postParent,err := strconv.ParseInt(r.FormValue("parent"),10,0)
	if err != nil{
		log.NewLog("",err)
		this.postJSONData(w,r,"",false)
		return
	}
	postStar,err := strconv.Atoi(r.FormValue("star"))
	if err != nil{
		log.NewLog("",err)
		this.postJSONData(w,r,"",false)
		return
	}
	postTitle := r.FormValue("title")
	postPage,err := strconv.Atoi(r.FormValue("page"))
	if err != nil{
		log.NewLog("",err)
		this.postJSONData(w,r,"",false)
		return
	}
	postMax,err := strconv.Atoi(r.FormValue("max"))
	if err != nil{
		log.NewLog("",err)
		this.postJSONData(w,r,"",false)
		return
	}
	postSort,err := strconv.Atoi(r.FormValue("sort"))
	if err != nil{
		log.NewLog("",err)
		this.postJSONData(w,r,"",false)
		return
	}
	postDesc := r.FormValue("desc")
	var postDescBool bool
	if postDesc == "true"{
		postDescBool = true
	}else{
		postDescBool = false
	}
	//get data
	collStatus,b := coll.GetStatus()
	if b == false{
		this.postJSONData(w,r,"",false)
		return
	}
	if collStatus["status"] == true{
		this.postJSONData(w,r,"",false)
		return
	}
	data,b := coll.ViewList(postCollName,postParent,postStar,postTitle,postPage,postMax,postSort,postDescBool)
	if b == false{
		this.postJSONData(w,r,"",false)
		return
	}
	this.postJSONData(w,r,data,true)
}

//Feedback center view content action
func (this *Handle) actionView(w http.ResponseWriter, r *http.Request) {
	if this.CheckLogin(w, r) == false {
		return
	}
	//Make sure that post / get is fine
	b := this.CheckURLPost(r)
	if b == false {
		return
	}
	this.UpdateLanguage()
	//get post
	postCollName := r.FormValue("coll")
	if postCollName == ""{
		this.PostText(w,r,"404 Error...")
		return
	}
	postID,err := strconv.ParseInt(r.FormValue("id"),10,0)
	if err != nil{
		log.NewLog("",err)
		this.PostText(w,r,"404 Error...")
		return
	}
	//get source src
	fileSrc := coll.View(postCollName,postID)
	if fileSrc == ""{
		this.PostText(w,r,"404 Error...")
		return
	}
	//get file data
	fileData,err := LoadFile(fileSrc)
	if err != nil{
		log.NewLog("",err)
		this.PostText(w,r,"404 Error...")
		return
	}
	_,err = w.Write(fileData)
	if err != nil{
		log.NewLog("",err)
		this.PostText(w,r,"404 Error...")
		return
	}
}