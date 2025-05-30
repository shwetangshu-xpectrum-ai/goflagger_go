package main

import (
    "html/template"
    "io/ioutil"
    "log"
    "net/http"
    "time"
    gofeatureflag "github.com/thomaspoignant/go-feature-flag"
    "github.com/thomaspoignant/go-feature-flag/ffcontext"
    "github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
    "gopkg.in/yaml.v2"
)

var (
    tmpl *template.Template
)

type FeatureFlag struct {
    Variations struct {
        Enabled  bool `yaml:"enabled"`
        Disabled bool `yaml:"disabled"`
    } `yaml:"variations"`
    Targeting []struct {
        Query     string `yaml:"query"`
        Variation string `yaml:"variation"`
    } `yaml:"targeting"`
    DefaultRule struct {
        Variation string `yaml:"variation"`
    } `yaml:"defaultRule"`
}

type FlagsConfig struct {
    BetaFeature        FeatureFlag `yaml:"beta-feature"`
    ExperimentalFeature FeatureFlag `yaml:"experimental-feature"`
}

type AdminFeature struct {
    Name            string
    DefaultVariation string
    TargetingRules  []struct {
        Query     string `yaml:"query"`
        Variation string `yaml:"variation"`
    }
}

func main() {
    // Initialize GO Feature Flag
    log.Println("Initializing GO Feature Flag...")
    err := gofeatureflag.Init(gofeatureflag.Config{
        PollingInterval: 3 * time.Second,
        Retriever: &fileretriever.Retriever{
            Path: "flags.yaml",
        },
    })
    if err != nil {
        log.Fatalf("Failed to initialize GO Feature Flag: %v", err)
    }
    defer gofeatureflag.Close()

    // Parse the HTML templates
    log.Println("Parsing HTML templates...")
    tmpl = template.Must(template.ParseFiles("templates/dashboard.html", "templates/admin.html"))

    // Set up the HTTP server
    log.Println("Starting HTTP server on :8080...")
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/dashboard", http.StatusMovedPermanently)
    })
    http.HandleFunc("/dashboard", dashboardHandler)
    http.HandleFunc("/admin", adminHandler)
    http.HandleFunc("/admin/update", adminUpdateHandler)
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
    
    log.Println("Server is ready! Visit http://localhost:8080/dashboard")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("user_id")
    if userID == "" {
        userID = "default"
    }
    
    ctx := ffcontext.NewEvaluationContext(userID)
    betaFeature, err := gofeatureflag.BoolVariation("beta-feature", ctx, false)
    if err != nil {
        log.Printf("Error evaluating beta-feature: %v", err)
    }
    
    experimentalFeature, err := gofeatureflag.BoolVariation("experimental-feature", ctx, false)
    if err != nil {
        log.Printf("Error evaluating experimental-feature: %v", err)
    }

    data := struct {
        BetaFeature         bool
        ExperimentalFeature bool
        UserID             string
    }{
        BetaFeature:         betaFeature,
        ExperimentalFeature: experimentalFeature,
        UserID:             userID,
    }
    
    if err := tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
        log.Printf("Error executing template: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
    // Read the current flags configuration
    data, err := ioutil.ReadFile("flags.yaml")
    if err != nil {
        log.Printf("Error reading flags.yaml: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    var config FlagsConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        log.Printf("Error parsing flags.yaml: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    features := []AdminFeature{
        {
            Name:            "beta-feature",
            DefaultVariation: config.BetaFeature.DefaultRule.Variation,
            TargetingRules:  config.BetaFeature.Targeting,
        },
        {
            Name:            "experimental-feature",
            DefaultVariation: config.ExperimentalFeature.DefaultRule.Variation,
            TargetingRules:  config.ExperimentalFeature.Targeting,
        },
    }

    if err := tmpl.ExecuteTemplate(w, "admin.html", map[string]interface{}{
        "Features": features,
    }); err != nil {
        log.Printf("Error executing template: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func adminUpdateHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Read the current configuration
    data, err := ioutil.ReadFile("flags.yaml")
    if err != nil {
        log.Printf("Error reading flags.yaml: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    var config FlagsConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        log.Printf("Error parsing flags.yaml: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Update beta-feature
    config.BetaFeature.DefaultRule.Variation = r.FormValue("beta-feature_default")
    config.BetaFeature.Targeting = make([]struct {
        Query     string `yaml:"query"`
        Variation string `yaml:"variation"`
    }, 0)
    
    queries := r.Form["beta-feature_query[]"]
    variations := r.Form["beta-feature_variation[]"]
    for i := range queries {
        if queries[i] != "" {
            config.BetaFeature.Targeting = append(config.BetaFeature.Targeting, struct {
                Query     string `yaml:"query"`
                Variation string `yaml:"variation"`
            }{
                Query:     queries[i],
                Variation: variations[i],
            })
        }
    }

    // Update experimental-feature
    config.ExperimentalFeature.DefaultRule.Variation = r.FormValue("experimental-feature_default")
    config.ExperimentalFeature.Targeting = make([]struct {
        Query     string `yaml:"query"`
        Variation string `yaml:"variation"`
    }, 0)
    
    queries = r.Form["experimental-feature_query[]"]
    variations = r.Form["experimental-feature_variation[]"]
    for i := range queries {
        if queries[i] != "" {
            config.ExperimentalFeature.Targeting = append(config.ExperimentalFeature.Targeting, struct {
                Query     string `yaml:"query"`
                Variation string `yaml:"variation"`
            }{
                Query:     queries[i],
                Variation: variations[i],
            })
        }
    }

    // Write the updated configuration
    newData, err := yaml.Marshal(config)
    if err != nil {
        log.Printf("Error marshaling config: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    if err := ioutil.WriteFile("flags.yaml", newData, 0644); err != nil {
        log.Printf("Error writing flags.yaml: %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // Redirect back to admin page
    http.Redirect(w, r, "/admin", http.StatusSeeOther)
} 