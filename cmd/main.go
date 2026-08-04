package main

import (
	"fmt"
	"log"

	"github.com/charmbracelet/lipgloss"
	"github.com/inestrivino/bonjour/internal/config"
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

func main() {
	bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).MarginBottom(1)

	fmt.Println(bannerStyle.Render(asciiTitle))

	cfg, err := config.LoadOrRunWizard()
	if err != nil {
		log.Fatalf("Error running wizard: %v", err)
	}

	fmt.Printf("Good Morning, %v!\n", cfg.User.Name)
}
