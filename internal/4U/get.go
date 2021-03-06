package fouru

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq" // Extension of the database/sql package
	"mynsb-api/internal/db"
	"mynsb-api/internal/quickerrors"
	"mynsb-api/internal/util"
	"net/http"
)


// RETRIEVAL FUNCTIONS

// getAll returns all 4U issues currently in our four_u table
func getAll(db *sql.DB) []Issue {
	issues, _ := performRequest(db, "SELECT * FROM four_u")
	return issues
}


// getBetween takes two times and returns all 4U articles published between those two times
func getBetween(times map[string]string, db *sql.DB) ([]Issue, error) {
	// Parse the dates
	start, parseErrOne := util.ParseDate(times["start"])
	end, parseErrTwo := util.ParseDate(times["end"])

	if parseErrOne != nil || parseErrTwo != nil {
		return []Issue{}, errors.New("could not parse date")
	}

	return performRequest(db, "SELECT * FROM four_u WHERE article_publish_date BETWEEN $1::TIMESTAMP AND $2::TIMESTAMP", start, end)
}












// UTILITY FUNCTIONS

// determineRequestType takes the sets of params fed to us by the user and determines what type of request they are sending
func determineRequestType(params map[string]string) string {

	// Determine type of request based on parsed parameters
	if util.IsSet(params["start"], params["end"]) {
		return "getBetween"
	}

	return "getAll"
}


// performRequest takes a simple request, parses it into an array and returns it
func performRequest(db *sql.DB, query string, args ...interface{}) ([]Issue, error) {
	// Array that will be returned
	var result []Issue

	// Perform the request
	res, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.Next() {
		article := Issue{}
		article.ReplaceWith(res)

		result = append(result, article)
	}

	return result, nil
}












// HTTP HANDLERS

// IssueRetrievalHandler deals with a request for a specific 4U Issue based off the parameters provided by the user
func IssueRetrievalHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

	// Start up database
	db.Conn("student")
	defer db.DB.Close()

	// Params sent to us by the user
	params := map[string]string{
		"start": r.URL.Query().Get("Start"),
		"end": r.URL.Query().Get("End"),
	}

	// Determine the request type for the incoming request
	requestType := determineRequestType(params)

	// Perform the request in regards to parameters
	switch requestType {

	case "getAll": // Request type = all

		// perform the request
		articles := getAll(db.DB)
		bytes, _ := json.Marshal(articles)
		util.HTTPResponse(200, "OK", string(bytes), "Response", w)
		break

	default: // Request type = between

		// perform the request
		res, err := getBetween(params, db.DB)
		if err != nil {
			quickerrors.InternalServerError(w)
			return
		}
		bytes, _ := json.Marshal(res)
		util.HTTPResponse(200, "Ok", string(bytes), "Response", w)
		break

	}
}