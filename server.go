package main

/**
** TODO - Remove uncessary imports here
**/
import (
	/* Package to write back to http.ResponseWriter */

	"fmt"
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
	type processReceiptResponse struct {
		id uint32
	}
	var requestBody processReceiptRequest
	err := json.Unmarshal([]byte(body), &requestBody)
	if err != nil {
		/*
		* Something went wrong while encoding it to json. Do proper error handling here
		 */
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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
		is_whole, err := regexp.MatchString(`[.].*[1-9].*`, requestBody.total)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if is_whole {
			points += 50
		}
		// Fractional part is one of either .0 | .25 | .75
		is_divisible_by_quarters, err := regexp.MatchString(`[.](25|75|0)[0]*$`, requestBody.total)
		if err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if is_divisible_by_quarters {
			points += 25
		}
		//5 points for every 2 items in the receipt
		points += uint32(5 * (len(requestBody.items) / 2))
		//For every item's trimmed description being a multiple of 3, add 0.2*trimmed_length points
		for _, item := range requestBody.items {
			if len(strings.Trim(item.shortDescription, " "))%3 == 0 {
				price, err := strconv.ParseFloat(item.price, 64)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				points += uint32(math.Round(price * 0.2))
			}
		}
		//6 points for day in date being odd
		day, err := strconv.ParseUint(strings.Split(requestBody.purcharseDate, "-")[2], 10, 32)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else if day%2 == 1 {
			points += 6
		}

		// 10 points while
		if "14:00" < requestBody.purchaseTime && requestBody.purchaseTime < "16:00" {
			points += 10
		}
		point_for_id[id] = points
		var processReceiptResponse processReceiptResponse
		processReceiptResponse.id = id
		response_body_bytes, err := json.Marshal(processReceiptResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, string(response_body_bytes))

	}

}

func get_points_handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")

}

func main() {
	http.HandleFunc("/receipts/process", process_receipt_handler)
	http.HandleFunc("/receipts/", get_points_handler)
	// Resolve using -> https://groups.google.com/g/golang-nuts/c/gQw-kxkoRGY?pli=1
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
