package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func UpdateLogPanel(panelX *tview.TextView) {
	// Don't clear, just update at the top by reading existing content
	gs := GetGameState()
	
	// Build status header - just HP bars and resources
	status := "[yellow]HAUNTED ROOM DEFENSE[white]\n\n"
	
	// HP Bars
	doorBar := DrawHPBar(gs.doorHP, gs.doorMaxHP, 30)
	status += fmt.Sprintf("[cyan]Door HP:[white] %s %d/%d\n", doorBar, gs.doorHP, gs.doorMaxHP)
	
	if gs.hunterActive {
		hunterBar := DrawHPBar(gs.hunterHP, gs.hunterMaxHP, 30)
		status += fmt.Sprintf("[red]Hunter HP:[white] %s %d/%d\n", hunterBar, gs.hunterHP, gs.hunterMaxHP)
		status += fmt.Sprintf("[red]Hunter Pos:[white] %d/10\n", gs.hunterPos)
	}
	
	status += "\n[yellow]RESOURCES[white]\n"
	status += fmt.Sprintf("[gold]Coins:[white] %d (+%.1f/s)\n", gs.coins, gs.coinsPerS)
	status += fmt.Sprintf("[cyan]Diamonds:[white] %d (+%.1f/s)\n\n", gs.diamonds, gs.diamPerS)
	status += "[yellow]RECENT LOGS[white]\n"
	
	// Get current text and find where logs start
	currentText := panelX.GetText(false)
	logsStart := -1
	searchStr := "RECENT LOGS"
	
	// Extract only the log messages
	if len(currentText) > 0 {
		if idx := findString(currentText, searchStr); idx >= 0 {
			// Find the position after the logs header line
			afterHeader := idx + len(searchStr)
			for afterHeader < len(currentText) && currentText[afterHeader] != '\n' {
				afterHeader++
			}
			if afterHeader < len(currentText) {
				afterHeader++ // Skip the newline
				logsStart = afterHeader
			}
		}
	}
	
	// Clear and write new content
	panelX.Clear()
	fmt.Fprint(panelX, status)
	
	// Restore existing logs
	if logsStart >= 0 && logsStart < len(currentText) {
		fmt.Fprint(panelX, currentText[logsStart:])
	}
}

func UpdateItemsPanel(panel *tview.TextView) {
	panel.Clear()
	
	gs := GetGameState()
	
	fmt.Fprintf(panel, "[yellow]YOUR ITEMS[white]\n\n")
	fmt.Fprintf(panel, "[cyan]Door:[white] Level %d (HP: %d)\n", gs.doorLevel, gs.doorMaxHP)
	fmt.Fprintf(panel, "[green]Beds:[white] %d (Coins/s: +%d)\n", gs.beds, gs.beds)
	fmt.Fprintf(panel, "[orange]Defense:[white] %d\n", gs.rooms[0].defense)
}

func UpdateShopPanel(panel *tview.TextView, selectedItem int) {
	panel.Clear()
	
	items := GetAvailableItems()
	
	fmt.Fprintf(panel, "[yellow]SHOP[white] [gray](U/D: navigate, I: buy)[white]\n\n")
	for i, item := range items {
		cursor := "  "
		if i == selectedItem {
			cursor = "[yellow]►[white] "
		}
		costColor := "[gold]"
		if item.costType == "diamond" {
			costColor = "[cyan]"
		}
		fmt.Fprintf(panel, "%s%s - %s%d %s[white]\n", cursor, item.name, costColor, item.cost, item.costType)
		fmt.Fprintf(panel, "   %s\n", item.description)
	}
}

func UpdateRoomDefensePanel(panel *tview.TextView) {
	panel.Clear()
	
	gs := GetGameState()
	room := gs.rooms[gs.currentRoom]
	
	fmt.Fprintf(panel, "[yellow]ROOM DEFENSE[white] [gray](←/→: switch room)[white]\n\n")
	fmt.Fprintf(panel, "[cyan]%s (%d/%d)[white]\n\n", room.name, gs.currentRoom+1, len(gs.rooms))
	fmt.Fprintf(panel, "[orange]Defense:[white] %d\n", room.defense)
	fmt.Fprintf(panel, "[gold]Coins/s:[white] %.1f\n", room.coinsPerS)
	fmt.Fprintf(panel, "[cyan]Diamonds/s:[white] %.1f\n", room.diamPerS)
}

func UpdateRoomItemsPanel(panel *tview.TextView) {
	panel.Clear()
	
	gs := GetGameState()
	room := gs.rooms[gs.currentRoom]
	
	fmt.Fprintf(panel, "[yellow]ROOM ITEMS[white]\n\n")
	
	if len(room.items) == 0 {
		fmt.Fprintf(panel, "[gray]No items[white]\n")
	} else {
		for _, item := range room.items {
			fmt.Fprintf(panel, "• %s\n", item)
		}
	}
	
	if gs.hunterActive {
		fmt.Fprintf(panel, "\n[red]⚠ HUNTER ACTIVE![white]\n")
		fmt.Fprintf(panel, "Position: %d/10\n", gs.hunterPos)
	}
}

func findString(text string, search string) int {
	for i := 0; i <= len(text)-len(search); i++ {
		if text[i:i+len(search)] == search {
			return i
		}
	}
	return -1
}
