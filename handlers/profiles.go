package handlers

import (
	"log"
	"net/http"
	"strconv"

	"llm-tournament/middleware"
	"llm-tournament/templates"
)

// ProfilesHandler handles the profiles page (backward compatible wrapper)
func ProfilesHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.Profiles(w, r)
}

// AddProfileHandler handles adding a profile (backward compatible wrapper)
func AddProfileHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.AddProfile(w, r)
}

// EditProfileHandler handles editing a profile (backward compatible wrapper)
func EditProfileHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.EditProfile(w, r)
}

// DeleteProfileHandler handles deleting a profile (backward compatible wrapper)
func DeleteProfileHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.DeleteProfile(w, r)
}

// ResetProfilesHandler handles resetting profiles (backward compatible wrapper)
func ResetProfilesHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ResetProfiles(w, r)
}

// Profiles handles the profiles page
func (h *Handler) Profiles(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling profiles page")
	searchQuery := r.FormValue("search_query")

	funcMap := templates.FuncMap

	pageName := "Profiles"
	profiles := h.DataStore.ReadProfiles()

	err := h.Renderer.Render(w, "profiles.html", funcMap, struct {
		PageName    string
		Profiles    []middleware.Profile
		SearchQuery string
	}{
		PageName:    pageName,
		Profiles:    profiles,
		SearchQuery: searchQuery,
	}, "templates/profiles.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	log.Println("Profiles page rendered successfully")
}

// AddProfile handles adding a profile
func (h *Handler) AddProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling add profile")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

	profiles := h.DataStore.ReadProfiles()
	profiles = append(profiles, middleware.Profile{Name: profileName, Description: profileDescription})
	err = h.DataStore.WriteProfiles(profiles)
	if err != nil {
		log.Printf("Error writing profiles: %v", err)
		http.Error(w, "Error writing profiles", http.StatusInternalServerError)
		return
	}
	log.Println("Profile added successfully")
	http.Redirect(w, r, "/profiles", http.StatusSeeOther)
}

// EditProfile handles editing a profile
func (h *Handler) EditProfile(w http.ResponseWriter, r *http.Request) {
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
		profiles := h.DataStore.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			funcMap := templates.FuncMap
			err := h.Renderer.Render(w, "edit_profile.html", funcMap, struct {
				Index   int
				Profile middleware.Profile
			}{
				Index:   index,
				Profile: profiles[index],
			}, "templates/edit_profile.html")
			if err != nil {
				log.Printf("Error rendering template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
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
		profiles := h.DataStore.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			oldProfileName := profiles[index].Name
			profiles[index].Name = editedProfileName
			profiles[index].Description = editedProfileDescription

			// Update prompts that reference this profile
			prompts := h.DataStore.ReadPrompts()
			for i := range prompts {
				if prompts[i].Profile == oldProfileName {
					prompts[i].Profile = editedProfileName
				}
			}
			err = h.DataStore.WritePrompts(prompts)
			if err != nil {
				log.Printf("Error updating prompts: %v", err)
				http.Error(w, "Error updating prompts", http.StatusInternalServerError)
				return
			}
		}
		err = h.DataStore.WriteProfiles(profiles)
		if err != nil {
			log.Printf("Error writing profiles: %v", err)
			http.Error(w, "Error writing profiles", http.StatusInternalServerError)
			return
		}
		log.Println("Profile edited successfully")
		http.Redirect(w, r, "/profiles", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// DeleteProfile handles deleting a profile
func (h *Handler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
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
		profiles := h.DataStore.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			funcMap := templates.FuncMap
			err := h.Renderer.Render(w, "delete_profile.html", funcMap, struct {
				Index   int
				Profile middleware.Profile
			}{
				Index:   index,
				Profile: profiles[index],
			}, "templates/delete_profile.html")
			if err != nil {
				log.Printf("Error rendering template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
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
		profiles := h.DataStore.ReadProfiles()
		if index >= 0 && index < len(profiles) {
			profiles = append(profiles[:index], profiles[index+1:]...)
		}
		err = h.DataStore.WriteProfiles(profiles)
		if err != nil {
			log.Printf("Error writing profiles: %v", err)
			http.Error(w, "Error writing profiles", http.StatusInternalServerError)
			return
		}
		log.Println("Profile deleted successfully")
		http.Redirect(w, r, "/profiles", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ResetProfiles handles resetting profiles
func (h *Handler) ResetProfiles(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset profiles")
	if r.Method == "GET" {
		if err := h.Renderer.RenderTemplateSimple(w, "reset_profiles.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		err := h.DataStore.WriteProfiles([]middleware.Profile{})
		if err != nil {
			log.Printf("Error writing profiles: %v", err)
			http.Error(w, "Error writing profiles", http.StatusInternalServerError)
			return
		}
		log.Println("Profiles reset successfully")
		http.Redirect(w, r, "/profiles", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
