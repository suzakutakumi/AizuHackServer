package AizuHackServer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
)

var client *firestore.Client
var ctx context.Context

func init() {
	conf := &firebase.Config{ProjectID: "aizuhack-353413"}

	ctx = context.Background()

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}
}

func CORS(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	fmt.Fprint(w, "Hello, World!")
}

func Data(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	// clientID, clientSecret, ok := r.BasicAuth()
	// if ok == false {
	// 	w.WriteHeader(http.StatusUnauthorized)
	// 	return
	// }
	// if clientID != os.Getenv("id") || clientSecret != os.Getenv("secret") {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	switch r.Method {
	case http.MethodGet:
		DataGet(w, r)
	case http.MethodPost:
		DataPost(w, r)
	case http.MethodPatch:
		DataPatch(w, r)
	case http.MethodDelete:
		DataDelete(w, r)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		fmt.Fprintln(w, "そんなメソッドねえよ")
	}
}

func DataGet(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	queryCol := query["collection"]
	if len(queryCol) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	col := queryCol[0]

	queryId := query["id"]
	queryKey := query["key"]
	if len(queryId) == 0 && len(queryKey) == 0 {
		docs, err := client.Collection(col).Documents(ctx).GetAll()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var data []map[string]interface{}
		for _, doc := range docs {
			data = append(data, doc.Data())
		}
		out, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, string(out))
	} else if len(queryId) == 1 && len(queryKey) == 1 {
		id := queryId[0]
		key := queryKey[0]
		doc, err := client.Collection(col).Doc(id).Get(ctx)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		data := doc.Data()
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, data[key])
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
func DataPost(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	col := query["collection"]
	if len(col) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	colName := col[0]

	bufbody := new(bytes.Buffer)
	bufbody.ReadFrom(r.Body)
	body := bufbody.Bytes()

	var jsonObj map[string]interface{}
	if err := json.Unmarshal(body, &jsonObj); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Println(jsonObj)

	doc, _, err := client.Collection(colName).Add(ctx, jsonObj)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	client.Collection(colName).Doc(doc.ID).Update(ctx, []firestore.Update{
		{
			Path:  "id",
			Value: doc.ID,
		},
	})
	w.WriteHeader(http.StatusOK)
}

func DataPatch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	queryCol := query["collection"]
	if len(queryCol) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	col := queryCol[0]

	queryId := query["id"]
	queryKey := query["key"]
	if len(queryId) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if len(queryKey) == 1 {
		id := queryId[0]
		key := queryKey[0]
		var body interface{}
		bufbody := new(bytes.Buffer)
		bufbody.ReadFrom(r.Body)
		body = string(bufbody.Bytes())
		_, err := client.Collection(col).Doc(id).Update(ctx, []firestore.Update{
			{
				Path:  key,
				Value: body,
			},
		})
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	if len(queryKey) == 0 {
		id := queryId[0]
		var data map[string]interface{}
		json.NewDecoder(r.Body).Decode(&data)
		var update []firestore.Update
		for k, v := range data {
			update = append(update, firestore.Update{
				Path:  k,
				Value: v,
			})
		}
		_, err := client.Collection(col).Doc(id).Update(ctx, update)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	return
}

func DataDelete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	col := query["collection"]
	ids := query["id"]
	if len(col) != 1 || len(ids) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	colName := col[0]
	id := ids[0]
	_, err := client.Collection(colName).Doc(id).Delete(ctx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
