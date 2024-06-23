package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
)

type RequestBody struct {
	Email   string `json:"email"`
	Lang    string `json:"lang"`
	Message string `json:"message"`
	Name    string `json:"name"`
	Subject string `json:"subject"`
}

type ErrorMessages map[string]map[string]string

var errorMessages = ErrorMessages{
	"DefaulErrorMessage": {
		"en": "An error occurred while processing the request!",
		"az": "Tələbi işlənən zaman səhv baş verdi!",
		"ru": "При обработке запроса произошла ошибка!",
	},
	"EmptyFields": {
		"en": "Please fill in all the fields!",
		"az": "Bütün tələb olunan sahələri doldurun!",
		"ru": "Заполните все обязательные поля!",
	},
	"NameTooLong": {
		"en": "The number of characters in the «Name» field must be no more than 50!",
		"az": "«Ad» sahəsindəki simvolların sayı 50-dən çox olmamalıdır!",
		"ru": "Количество символов в поле «Имя» не должно превышать 50!",
	},
	"EmailFieldIsRequired": {
		"en": "The «Email» field is required!",
		"az": "«Email» sahəsi tələb olunur!",
		"ru": "Поле «Email» обязательно!",
	},
	"InvalidEmail": {
		"en": "Please enter a valid email address!",
		"az": "Zəhmət olmasa düzgün bir email ünvanı daxil edin!",
		"ru": "Пожалуйста, введите действительный адрес электронной почты!",
	},
	"SubjectTooLong": {
		"en": "The number of characters in the «Subject» field must be no more than 50!",
		"az": "«Mövzu» sahəsindəki simvolların sayı 50-dən çox olmamalıdır!",
		"ru": "Количество символов в поле «Тема» не должно превышать 50!",
	},
	"MessageFieldIsRequired": {
		"en": "The «Message» field is required!",
		"az": "«Mesaj» sahəsi tələb olunur!",
		"ru": "Поле «Сообщение» обязательно!",
	},
	"MessageTooShort": {
		"en": "The message must be at least 1 character long!",
		"az": "Mesaj ən az 1 simvol uzunluğunda olmalıdır!",
		"ru": "Сообщение должно быть длиной не менее 1 символа!",
	},
}

func getErrorMessage(errorType string, lang string) string {
	// Check if the error type exists
	if messages, exists := errorMessages[errorType]; exists {
		// Check if the language exists for the given error type
		if message, exists := messages[lang]; exists {
			return message
		}
		// Default to English if the language is not recognized
		return messages["en"]
	}
	// Default error message if the error type is not recognized
	return "An unknown error occurred!"
}

func validateEmail(email string) bool {
	// Regular expression for validating an Email
	var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	emailLength := len(email)
	if !emailRegex.MatchString(email) || emailLength < 7 || emailLength > 50 {
		return false
	}

	return true
}

func (request RequestBody) validateResponse() (bool, error) {
	var errorType string

	switch {
	case len(request.Email) == 0 || len(request.Message) == 0:
		errorType = "EmptyFields"
	case len(request.Name) > 50:
		errorType = "NameTooLong"
	case len(request.Email) == 0:
		errorType = "EmailFieldIsRequired"
	case !validateEmail(request.Email):
		errorType = "InvalidEmail"
	case len(request.Subject) > 0:
		errorType = "SubjectTooLong"
	case len(request.Message) == 0:
		errorType = "MessageFieldIsRequired"
	case len(request.Message) < 1:
		errorType = "MessageTooShort"
	}

	if errorType != "" {
		err := getErrorMessage(errorType, request.Lang)
		return false, errors.New(err)
	}

	return true, nil
}

// CORSHandler is a middleware that adds the necessary CORS headers to the response
func CORSHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8899")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	var reponse_message string
	if r.Method == http.MethodPost {
		// Read the body of the request
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Unable to read body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Unmarshal the JSON data into the RequestBody struct
		var request RequestBody
		err = json.Unmarshal(body, &request)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		isValid, err := request.validateResponse()
		if !isValid && err != nil {
			reponse_message = err.Error()
		} else {
			switch request.Lang {
			case "en":
				reponse_message = "Your message has been sent successfully!"
			case "az":
				reponse_message = "Mesajınız uğurla göndərildi!"
			case "ru":
				reponse_message = "Ваше сообщение успешно отправлено!"
			}
		}

		w.Write([]byte(reponse_message))
	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	http.ListenAndServe("", CORSHandler(mux))
}
