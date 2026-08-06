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
 __                                                    
/\ \                        __                         
\ \ \____    ___     ___   /\_\    ___   __  __  _ __  
 \ \ '` + "`" + `__` + "`" + `\  / __` + "`" + `\ /' _ ` + "`" + `\ \/\ \  / __` + "`" + `\/\ \/\ \/\` + "`" + `'__\
  \ \ \L\ \/\ \L\ \/\ \/\ \ \ \ \/\ \L\ \ \ \_\ \ \ \/ 
   \ \_,__/\ \____/\ \_\ \_\_\ \ \ \____/\ \____/\ \_\ 
    \/___/  \/___/  \/_/\/_/\ \_\ \/___/  \/___/  \/_/ 
                           \ \____/                    
                            \/___/`

// greeting takes in a Config type object and a theme type object, from which it renders a greeting based on the user's name, time of day, and theme
func greeting(cfg *config.Config, theme *ui.Theme) {
	greeting := "morning"
	hour := time.Now().Hour()

	if hour >= 12 && hour < 18 {
		greeting = "afternoon"
	} else if hour >= 18 {
		greeting = "evening"
	}

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
