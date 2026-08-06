package main

import (
	"fmt"
	"log"
	"time"

	"github.com/inestrivino/bonjour/internal/config"
	"github.com/inestrivino/bonjour/internal/modules/quotes"
	"github.com/inestrivino/bonjour/internal/ui"
)

const asciiTitle = `
 ____                                                  
/\  _` + "`" + `\                     __                         
\ \ \L\ \    ___     ___   /\_\    ___   __  __  _ __  
 \ \  _ <'  / __` + "`" + `\ /' _ ` + "`" + `\ \/\ \  / __` + "`" + `\/\ \/\ \/\` + "`" + `'__\
  \ \ \L\ \/\ \L\ \/\ \/\ \ \ \ \/\ \L\ \ \ \_\ \ \ \/ 
   \ \____/\ \____/\ \_\ \_\_\ \ \ \____/\ \____/\ \_\ 
    \/___/  \/___/  \/_/\/_/\ \_\ \/___/  \/___/  \/_/ 
                           \ \____/                    
                            \/___/`

func greeting(cfg *config.Config, theme *ui.Theme) {
	greeting := "morning"
	hour := time.Now().Hour()

	if hour >= 12 && hour < 18 {
		greeting = "afternoon"
	} else if hour >= 18 {
		greeting = "evening"
	}

	// Use theme styling for the greeting text instead of raw stdout
	fmt.Println(theme.Subtitle.Render(fmt.Sprintf("Good %s, %s!\n", greeting, cfg.User.Name)))
}

func main() {
	// Load or create configuration
	cfg, err := config.LoadOrRunWizard()
	if err != nil {
		log.Fatalf("Error running wizard: %v", err)
	}

	// Initialize the theme from the configuration
	theme := ui.NewTheme(cfg.Dashboard.Theme)

	// Render the banner
	fmt.Println(theme.Banner.Render(asciiTitle))

	// Render greeting
	greeting(cfg, theme)

	// Render the active modules
	fmt.Println(quotes.RenderQuote(theme))
}
