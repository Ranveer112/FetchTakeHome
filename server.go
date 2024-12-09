package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

/*
* A hashtable of id as key and numeric points as value
 */
var points_for_ids map[uint32]uint32
var smallest_unused_id uint32

func process_receipt_handler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Improper request body", http.StatusBadRequest)
		return
	}
	type item struct {
		ShortDescription string
		Price            string
	}
	type processReceiptRequest struct {
		Retailer     string
		PurchaseDate string
		PurchaseTime string
		Items        []item
		Total        string
	}
	type processReceiptResponse struct {
		Id uint32
	}
	var requestBody processReceiptRequest
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		/*
		* Something went wrong while encoding it to json. Do proper error handling here
		 */
		http.Error(w, "Improper request body.", http.StatusBadRequest)
		return
	} else {

		var id = smallest_unused_id
		smallest_unused_id++
		var points uint32
		//1 point for every alphanumeric character
		for _, element := range requestBody.Retailer {
			if unicode.IsNumber(element) || unicode.IsLetter(element) {
				points++
			}
		}
		//50 points for total being a whole number
		is_fractional, err := regexp.MatchString(`[.].*[1-9].*`, requestBody.Total)
		if err != nil {
			http.Error(w, "Improper request body - Total value is improper", http.StatusBadRequest)
			return
		} else if !is_fractional {
			points += 50
		}
		// Fractional part is one of either .0 | .25 | .75
		is_divisible_by_quarters, err := regexp.MatchString(`[.](25|75|0)[0]*$`, requestBody.Total)
		if err != nil {

			http.Error(w, "Improper request body - Total value is improper", http.StatusBadRequest)
			return
		} else if is_divisible_by_quarters {
			points += 25
		}
		//5 points for every 2 items in the receipt
		points += uint32(5 * (len(requestBody.Items) / 2))
		//For every item's trimmed description being a multiple of 3, add 0.2*(item's price) points
		for _, item := range requestBody.Items {
			if len(strings.Trim(item.ShortDescription, " "))%3 == 0 {
				price, err := strconv.ParseFloat(item.Price, 64)
				if err != nil {
					http.Error(w, "Improper request body - Price value of items is improper", http.StatusBadRequest)
					return
				}
				points += uint32(math.Ceil(price * 0.2))
			}
		}
		//6 points for day in date being odd
		day, err := strconv.ParseUint(strings.Split(requestBody.PurchaseDate, "-")[2], 10, 32)
		if err != nil {
			http.Error(w, "Improper request body - Purchase date value is improper", http.StatusBadRequest)
			return
		} else if day%2 == 1 {
			points += 6
		}

		// 10 points for PurchaseTime being after 2:00PM and before 4:00PM
		if "14:00" < requestBody.PurchaseTime && requestBody.PurchaseTime < "16:00" {
			points += 10
		}
		points_for_ids[id] = points
		var processReceiptResponse processReceiptResponse
		processReceiptResponse.Id = id
		response_body_bytes, err := json.Marshal(processReceiptResponse)
		if err != nil {
			http.Error(w, "Something went wrong on server side.", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(response_body_bytes))
	}

}

func get_points_handler(w http.ResponseWriter, r *http.Request) {
	var input_id = r.PathValue("id")
	id, err := strconv.ParseUint(input_id, 10, 32)
	if err != nil {
		http.Error(w, "ID in the request is improper. Should be a base 10 integer", http.StatusBadRequest)
		return
	}
	point_for_id, id_exists := points_for_ids[uint32(id)]
	if !id_exists {
		http.Error(w, "No receipt with ID found", http.StatusBadRequest)
	} else {
		fmt.Fprintf(w, strconv.FormatUint(uint64(point_for_id), 10))
	}

}
func init_database() {
	points_for_ids = make(map[uint32]uint32)
	smallest_unused_id = 0
}

func main() {
	init_database()
	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()
	fmt.Printf("Server listening on port %s\n", *port)
	http.HandleFunc("POST /receipts/process", process_receipt_handler)
	http.HandleFunc("GET /receipts/{id}", get_points_handler)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
