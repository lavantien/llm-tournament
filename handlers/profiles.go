package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"llm-tournament/middleware"
)

// Handle profiles page
func ProfilesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling profiles page")
	searchQuery := r.FormValue("search_query")

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"tolower": strings.ToLower,
		"contains": strings.Contains,
	}
	funcMap["json"] = func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	}

	t, err := template.New("profiles.html").Funcs(funcMap).ParseFiles("templates/profiles.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	if t == nil {
		log.Println("Error parsing template")
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, struct {
		Profiles    []middleware.Profile
		SearchQuery string
	}{
		Profiles:    profiles,
		SearchQuery: searchQuery,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
	log.Println("Profiles page rendered successfully")
}

// Handle add profile
func AddProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling add profile")
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	profileName := r.Form.Get("profile_name")
	profileDescription := r.Form.Get("profile_description")
	if profileName == "" {
		log.Println("Profile name cannot be empty")
		http.Error(w, "Profile name cannot be empty", http.StatusBadRequest)
		return
	}

	profiles := middleware.ReadProfiles()
	profiles = append(profiles, middleware.Profile{Name: profileName, Description: profileDescription})
	err = middleware.WriteProfiles(profiles)
	if err != nil {
		log.Printf("Error writing profiles: %v", err)
		http.Error(w, "Error writing profiles", http.StatusInternalServerError)
		return
	}
	log.Println("Profile added successfully")
	http.Redirect(w, r, "/profiles", http.StatusSeeOther)
}

// Handle edit profile
func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling edit profile")
	if r.Method == "GET" {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		profiles := middleware.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			funcMap := template.FuncMap{
				"markdown": func(text string) template.HTML {
					unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
					html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
					return template.HTML(html)
				},
			}
			t, err := template.New("edit_profile.html").Funcs(funcMap).ParseFiles("templates/edit_profile.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index   int
				Profile middleware.Profile
			}{
				Index:   index,
				Profile: profiles[index],
			})
			if err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Error executing template", http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		editedProfileName := r.Form.Get("profile_name")
		editedProfileDescription := r.Form.Get("profile_description")
		if editedProfileName == "" {
			log.Println("Profile name cannot be empty")
			http.Error(w, "Profile name cannot be empty", http.StatusBadRequest)
			return
		}
		profiles := middleware.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			profiles[index].Name = editedProfileName
			profiles[index].Description = editedProfileDescription
		}
		err = middleware.WriteProfiles(profiles)
		if err != nil {
			log.Printf("Error writing profiles: %v", err)
			http.Error(w, "Error writing profiles", http.StatusInternalServerError)
			return
		}
		log.Println("Profile edited successfully")
		http.Redirect(w, r, "/profiles", http.StatusSeeOther)
	}
}

// Handle delete profile
func DeleteProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete profile")
	if r.Method == "GET" {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		profiles := middleware.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			funcMap := template.FuncMap{
				"markdown": func(text string) template.HTML {
					unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
					html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
					return template.HTML(html)
				},
			}
			t, err := template.New("delete_profile.html").Funcs(funcMap).ParseFiles("templates/delete_profile.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index   int
				Profile middleware.Profile
			}{
				Index:   index,
				Profile: profiles[index],
			})
			if err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Error executing template", http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		profiles := middleware.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			profiles = append(profiles[:index], profiles[index+1:]...)
		}
		err = middleware.WriteProfiles(profiles)
		if err != nil {
			log.Printf("Error writing profiles: %v", err)
			http.Error(w, "Error writing profiles", http.StatusInternalServerError)
			return
		}
		log.Println("Profile deleted successfully")
		http.Redirect(w, r, "/profiles", http.StatusSeeOther)
	}
}

// Handle reset profiles
func ResetProfilesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset profiles")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/reset_profiles.html")
		if err != nil {
			http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		err := middleware.WriteProfiles([]middleware.Profile{})
		if err != nil {
			log.Printf("Error writing profiles: %v", err)
			http.Error(w, "Error writing profiles", http.StatusInternalServerError)
			return
		}
		log.Println("Profiles reset successfully")
		http.Redirect(w, r, "/profiles", http.StatusSeeOther)
	}
}
