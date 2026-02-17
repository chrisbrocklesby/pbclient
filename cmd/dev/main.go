package main

import (
	"log"
	"net/http"
	"strings"

	pbclient "app"
)

type Post struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Rich    string `json:"rich"`
}

type User struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Avatar   string `json:"avatar"`
}

func main() {
	// Init default client with logger + base URL.
	_, err := pbclient.New(pbclient.Config{
		BaseURL: "http://127.0.0.1:8090",
		Logger:  log.Default(),
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := pbclient.LoginSuperAdmin("user@example.com", "password1234"); err != nil {
		log.Fatal(err)
	}

	posts := pbclient.Collection[Post]("posts")
	users := pbclient.Collection[User]("users")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		list, err := posts.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Posts:\n"))
		for _, post := range list.Items {
			_, _ = w.Write([]byte("- " + post.Title + "\n"))
			if post.Rich != "" {
				_, _ = w.Write([]byte("  " + post.Rich + "\n"))
			}
		}
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			http.NotFound(w, r)
			return
		}
		list, err := users.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Users:\n"))
		for _, user := range list.Items {
			_, _ = w.Write([]byte("- " + user.Name + " (" + user.Email + ")\n"))
			if user.Avatar != "" {
				_, _ = w.Write([]byte("  Avatar: " + user.Avatar + "\n"))
			}
		}
	})

	http.HandleFunc("/create", func(w http.ResponseWriter, r *http.Request) {
		// if r.Method != http.MethodPost {
		// 	w.Header().Set("Allow", http.MethodPost)
		// 	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		// 	return
		// }
		_, err := posts.Create(map[string]any{
			"title":   "Hello World 123",
			"content": "This is a test post",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Post created!"))
	})

	// net/http default mux does not support /view/{id} patterns; parse manually.
	http.HandleFunc("/view/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/view/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		post, err := posts.Get(id)
		if err != nil {
			http.Error(w, "Post not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Title: " + post.Title + "\nContent: " + post.Content))
	})

	http.HandleFunc("/update/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/update/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		_, err := posts.Update(id, map[string]any{
			"title":   "Updated Title",
			"content": "Updated content",
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Post updated!"))
	})

	http.HandleFunc("/delete/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/delete/")
		if id == "" {
			http.NotFound(w, r)
			return
		}
		if err := posts.Delete(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("Post deleted!"))
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
