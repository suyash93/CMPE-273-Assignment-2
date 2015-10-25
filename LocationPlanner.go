package main 
import (
"fmt"
"math/rand"
"strings"
"net/http"
"github.com/julienschmidt/httprouter"
"encoding/json"
"io/ioutil"
"os"
"log"
"strconv"
"gopkg.in/mgo.v2"
"gopkg.in/mgo.v2/bson"
)
type (UserController struct {
	session *mgo.Session
	})
type Request struct {
	Name string `json:"name"`
	Address string `json:"address"`
	City string `json:"city" `
	State string `json:"state"`
	Zip string `json:"zip"`
}
	
type Response struct {
	Address    string `json:"address" bson:"address"`
	City       string `json:"city" bson:"city"`
	Coordinate struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
	} `json:"coordinate" bson:"coordinate"`
	ID    int   `json:"id" bson:"id"`
	Name  string `json:"name" bson:"name"`
	State string `json:"state" bson:"state"`
	Zip   string `json:"zip" bson:"zip"`
}
//JSON struct from GooGle Maps Api
type GoogleMaps struct {
	Results []struct {
		AddressComponents []struct {
			LongName string `json:"long_name"`
			ShortName string `json:"short_name"`
			Types []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string `json:"place_id"`
		Types []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

//Function for HTTP POST
func  PostRequest(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	decoder:= json.NewDecoder(r.Body)

	var u Request
	err:= decoder.Decode(&u)
	if err!=nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var c Response
    c.Name= u.Name
    c.Address= u.Address
    c.City= u.City
    c.State= u.State
    c.Zip= u.Zip 
    c.ID=rand.Intn(10000000)
    var fulladdress string 
    fulladdress= c.Address+" "+c.City
    latresponse := GetLatitude(fulladdress)
    longresponse := GetLongitude(fulladdress)
    c.Coordinate.Lat= latresponse
    c.Coordinate.Lng= longresponse
    sess:=getSession();
    collection:= sess.DB("trip-planner").C("locations")
    e:= collection.Insert(c)
    if e!=nil {
    	panic(e)
    }

	uj,_ := json.Marshal(c)
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
	fmt.Fprintf(w, "%s", uj)
}
//Function to obtain Latitude coordinates for a user's location
func GetLatitude(fulladdress string) float64{
	var Ad GoogleMaps
	var lat64 float64
	Baseur:= "http://maps.google.com/maps/api/geocode/json?address="
	Addur:= fulladdress
	Urlf:= Baseur + Addur
	Urlf = strings.Replace(Urlf," ","%20",-1)
    	apiRes, err:= http.Get(Urlf)
	if err!=nil {
		fmt.Printf("error occurred")
		fmt.Printf("%s", err)
		os.Exit(1)
	}else{
		defer apiRes.Body.Close()
		contents, err:= ioutil.ReadAll(apiRes.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
	   err= json.Unmarshal(contents, &Ad)
	   if err!=nil {
	   	fmt.Println("her is the error")
	   	fmt.Printf("%s", err)
	   	os.Exit(1)
	   }
	 lat64 = Ad.Results[0].Geometry.Location.Lat
	}
	 return lat64
}

//Function to access Longitude coordinates for a user's location
func GetLongitude(fulladdress string) float64{
	var s GoogleMaps
	var long64 float64
	Baseurl:= "http://maps.google.com/maps/api/geocode/json?address="
	Addurl:= fulladdress
	Url:= Baseurl + Addurl
	Url = strings.Replace(Url," ","%20",-1)
	apiResponse, err:= http.Get(Url)
	if err!=nil {
		fmt.Printf("error occurred")
		fmt.Printf("%s", err)
		os.Exit(1)
	}else{
		defer apiResponse.Body.Close()
		contents, err:= ioutil.ReadAll(apiResponse.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		err= json.Unmarshal(contents, &s)
		if err!=nil {
			fmt.Println("Here is the error from longitude")
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		long64=s.Results[0].Geometry.Location.Lng		
}
return long64
}

//Function to obtain HTTP GET Request
func  GetRequest(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	 var updatedmsg Response
	 a:= p.ByName("id")
	 ac,_ := strconv.Atoi(a)
	   sess:=getSession();
  er := sess.DB("trip-planner").C("locations").Find(bson.M{"id": ac}).One(&updatedmsg)
  if er!=nil {
  	panic(er)
  }
	uj,_ := json.Marshal(updatedmsg)
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
	fmt.Fprintf(w, "%s", uj)
}

//Function to obtain HTTP DELETE Request
func  DeleteRequest(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	 a:= p.ByName("id")
	 ac,_ := strconv.Atoi(a)
	   sess:=getSession();
  er := sess.DB("trip-planner").C("locations").Remove(bson.M{"id": ac})
  if er!=nil {
  	panic(er)
  }
	w.WriteHeader(200)
}

//Function to access HTTP PUT Request
func  PutRequest(w http.ResponseWriter, res *http.Request, p httprouter.Params){
	decoder:= json.NewDecoder(res.Body)

	var r Request
	err:= decoder.Decode(&r)
	if err!=nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var oldupdatingmsg Response
	var updatingmsg Response
	init:= p.ByName("id")
	abs, _:= strconv.Atoi(init)
	newsession:= getSession();
	errors:= newsession.DB("trip-planner").C("locations").Find(bson.M{"id": abs}).One(&oldupdatingmsg)
	if errors!=nil {
		panic(errors)
	}
		updatingmsg.Name= r.Name
	    updatingmsg.Address= r.Address
	    updatingmsg.City= r.City
	    updatingmsg.Zip= r.Zip
	    updatingmsg.State= r.State
        updatingmsg.ID= abs
	var updateaddress string
	updateaddress= updatingmsg.Address+updatingmsg.City
	updatelatresp:= GetLatitude(updateaddress)
	updatelongresp:= GetLongitude(updateaddress)
	updatingmsg.Coordinate.Lat= updatelatresp
	updatingmsg.Coordinate.Lng= updatelongresp
	collec:= newsession.DB("trip-planner").C("locations")
    ef:= collec.Update(oldupdatingmsg,updatingmsg)
    if ef!=nil {
    	panic(ef)
    }
    updatejson,_ := json.Marshal(updatingmsg)
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(201)
	fmt.Fprintf(w, "%s", updatejson)
}


func main() {
	router:= httprouter.New()
	router.GET("/:id", GetRequest)
	router.POST("/", PostRequest)
	router.DELETE("/:id", DeleteRequest)
	router.PUT("/:id", PutRequest)
	log.Fatal(http.ListenAndServe(":6547", router))
}

//Function to connect to MongoDB
func getSession() *mgo.Session {
	connect, err:= mgo.Dial("mongodb://suyash:123@ds031531.mongolab.com:31531/trip-planner")
	if err!= nil {
			panic(err)
		}	
		return connect
}