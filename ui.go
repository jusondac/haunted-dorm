package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func UpdateLogPanel(panelX *tview.TextView) {
	// Don't clear, just update at the top by reading existing content
	gs := GetGameState()
	
	// Build status header - just title and hunter info if active
	status := "[yellow]HAUNTED ROOM DEFENSE[white]\n\n"
	
	if gs.hunterActive {
		hunterBar := DrawHPBar(gs.hunterHP, gs.hunterMaxHP, 30)
		status += fmt.Sprintf("[red]Hunter HP:[white] %s %d/%d\n", hunterBar, gs.hunterHP, gs.hunterMaxHP)
		status += fmt.Sprintf("[red]Hunter Pos:[white] %d/10\n\n", gs.hunterPos)
	}
	
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
	
	// Live resources
	fmt.Fprintf(panel, "[gold]Coins:[white] %d (+%.1f/s)\n", gs.coins, gs.coinsPerS)
	fmt.Fprintf(panel, "[cyan]Diamonds:[white] %d (+%.1f/s)\n\n", gs.diamonds, gs.diamPerS)
	
	// Items
	fmt.Fprintf(panel, "[cyan]Door:[white] Level %d (HP: %d)\n", gs.doorLevel, gs.doorMaxHP)
	if gs.bedLevel > 0 {
		fmt.Fprintf(panel, "[green]Bed:[white] Level %d (+%.0f coins/s)\n", gs.bedLevel, gs.coinsPerS)
	}
	if gs.playboxLevel > 0 {
		fmt.Fprintf(panel, "[magenta]Playbox:[white] Level %d (+%.0f diamonds/s)\n", gs.playboxLevel, gs.diamPerS)
	}
	fmt.Fprintf(panel, "[orange]Defense:[white] %d\n", gs.playerMaxDefense)
}

func UpdateShopPanel(panel *tview.TextView, selectedItem int) {
	panel.Clear()
	
	items := GetAvailableItems()
	
	fmt.Fprintf(panel, "[yellow]SHOP[white] [gray](↑/↓: navigate, I: buy)[white]\n\n")
	for i, item := range items {
		color := GetItemColor(item)
		
		// Highlight selected item with background
		if i == selectedItem {
			fmt.Fprintf(panel, "[black:white]%s%s[white:-]\n", color, item.name)
		} else {
			fmt.Fprintf(panel, "%s%s[white]\n", color, item.name)
		}
		
		// Show cost
		costStr := ""
		if item.costCoins > 0 && item.costDiamonds > 0 {
			costStr = fmt.Sprintf("[gold]%dc[white] + [cyan]%dd[white]", item.costCoins, item.costDiamonds)
		} else if item.costCoins > 0 {
			costStr = fmt.Sprintf("[gold]%dc[white]", item.costCoins)
		} else if item.costDiamonds > 0 {
			costStr = fmt.Sprintf("[cyan]%dd[white]", item.costDiamonds)
		}
		
		fmt.Fprintf(panel, "  Cost: %s\n", costStr)
		fmt.Fprintf(panel, "  %s\n", item.description)
	}
}

func UpdateRoomDefensePanel(panel *tview.TextView) {
	panel.Clear()
	
	gs := GetGameState()
	room := gs.rooms[gs.currentRoom]
	
	fmt.Fprintf(panel, "[yellow]DREAMERS[white]\n\n")
	
	// Show player first
	playerBar := DrawHPBar(gs.playerDefense, gs.playerMaxDefense, 20)
	fmt.Fprintf(panel, "[green]You:[white] %s %d/%d\n", playerBar, gs.playerDefense, gs.playerMaxDefense)
	
	// Show AI characters
	for _, char := range room.characters {
		charBar := DrawHPBar(char.defense, char.maxDefense, 20)
		fmt.Fprintf(panel, "[cyan]%s:[white] %s %d/%d\n", char.name, charBar, char.defense, char.maxDefense)
	}
	
	fmt.Fprintf(panel, "\n[yellow]RESOURCES[white]\n")
	fmt.Fprintf(panel, "[gold]Coins:[white] %d (+%.1f/s)\n", gs.coins, gs.coinsPerS)
	fmt.Fprintf(panel, "[cyan]Diamonds:[white] %d (+%.1f/s)\n", gs.diamonds, gs.diamPerS)
}

func UpdateRoomItemsPanel(panel *tview.TextView) {
	panel.Clear()
	
	gs := GetGameState()
	
	fmt.Fprintf(panel, "[yellow]ROOM ITEMS[white]\n\n")
	
	// Show door HP
	doorBar := DrawHPBar(gs.doorHP, gs.doorMaxHP, 20)
	fmt.Fprintf(panel, "[cyan]Door:[white] %s %d/%d\n\n", doorBar, gs.doorHP, gs.doorMaxHP)
	
	if len(gs.rooms[gs.currentRoom].items) == 0 {
		fmt.Fprintf(panel, "[gray]No items[white]\n")
	} else {
		for _, item := range gs.rooms[gs.currentRoom].items {
			fmt.Fprintf(panel, "• %s\n", item)
		}
	}
	
	if gs.hunterActive {
		fmt.Fprintf(panel, "\n[red]⚠ HUNTER ACTIVE![white]\n")
		fmt.Fprintf(panel, "Position: %d/10\n", gs.hunterPos)
		
		hunterBar := DrawHPBar(gs.hunterHP, gs.hunterMaxHP, 20)
		fmt.Fprintf(panel, "[red]HP:[white] %s %d/%d\n", hunterBar, gs.hunterHP, gs.hunterMaxHP)
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
