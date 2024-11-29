package main

/**
** TODO - Remove uncessary imports here
**/
import (
	/* Package to write back to http.ResponseWriter */

	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	/* Package for http */
	"net/http"
	/* Package to write error message */
	"log"
	/* Package to encode/decode JSON */
	"encoding/json"
)

/*
* A hashtable of id as key and numeric points as value
 */
var point_for_id map[uint32]uint32
var smallest_unused_id uint32

func process_receipt_handler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	type item struct {
		shortDescription string
		price            string
	}
	type processReceiptRequest struct {
		retailer      string
		purcharseDate string
		purchaseTime  string
		items         []item
		total         string
	}
	var requestBody processReceiptRequest
	err := json.Unmarshal([]byte(body), &requestBody)
	if err != nil {
		/*
		* Something went wrong while encoding it to json. Do proper error handling here
		 */

	} else {
		/*
		*Assume the request has been validated.
		 */
		var id = smallest_unused_id
		smallest_unused_id++
		var points uint32
		//1 point for every alphanumeric character
		for _, element := range requestBody.retailer {
			if unicode.IsNumber(element) || unicode.IsLetter(element) {
				points++
			}
		}
		//50 points for total being a whole number
		matched, err := regexp.MatchString(`[.].*[1-9].*`, requestBody.total)
		if err != nil {
			/*
			* Something went wrong while pattern matching. Do proper error handling here
			 */
		} else {
			if matched {
				points += 50
			}
			/*
				TODO - Implement 0.25 multiple point logic
			*/
			//5 points for every 2 items in the receipt
			points += uint32(5 * (len(requestBody.items) / 2))
			//For every item's trimmed description being a multiple of 3, add 0.2*trimmed_length points
			for _, item := range requestBody.items {
				if len(strings.Trim(item.shortDescription, " "))%3 == 0 {
					price, err := strconv.ParseFloat(item.price, 64)
					if err != nil {
						/*
						* Something went wrong while price conversion to float. Do proper error handling here
						 */
					}
					points += uint32(math.Round(price * 0.2))
				}
			}
			//6 points for day in date being odd
			day, err := strconv.ParseUint(strings.Split(requestBody.purcharseDate, "-")[2], 10, 32)
			if err != nil {
				/*
					Something went wrong while parsing the day portion of the date, do proper handling here
				*/
			} else if day%2 == 1 {
				points += 6
			}
			purchaseTime := strings.Split(requestBody.purchaseTime, ":")
			/*
				TODO - Add time based point calculation
			*/
			point_for_id[id] = points

		}
	}

}

func get_points_handler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	http.HandleFunc("/receipts/process", process_receipt_handler)
	http.HandleFunc("/receipts/", get_points_handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/*
v2 part of this project
Use https://github.com/getkin/kin-openapi maybe

func main() {
	ctx := context.Background()
	loader := &openapi3.Loader{Context: ctx, IsExternalRefsAllowed: true}
	doc, _ := loader.LoadFromFile(".../My-OpenAPIv3-API.yml")
	// Validate document
	_ = doc.Validate(ctx)
	router, _ := gorillamux.NewRouter(doc)
	httpReq, _ := http.NewRequest(http.MethodGet, "/items", nil)

	// Find route
	route, pathParams, _ := router.FindRoute(httpReq)

	// Validate request
	requestValidationInput := &openapi3filter.RequestValidationInput{
		Request:    httpReq,
		PathParams: pathParams,
		Route:      route,
	}
	_ = openapi3filter.ValidateRequest(ctx, requestValidationInput)


    http.HandleFunc("/receipts/process", process_receipt_handler)
	http.HandleFunc("/receipts/", get_points_handler)

	responseHeaders := http.Header{"Content-Type": []string{"application/json"}}
	responseCode := 200
	responseBody := []byte(`{}`)

	// Validate response
	responseValidationInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: requestValidationInput,
		Status:                 responseCode,
		Header:                 responseHeaders,
	}
	responseValidationInput.SetBodyBytes(responseBody)
	_ = openapi3filter.ValidateResponse(ctx, responseValidationInput)
}
*/
