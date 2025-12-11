package main

import (
	"fmt"

	"github.com/rivo/tview"
)

func UpdateLogPanel(panelX *tview.TextView) {
	// Don't clear, just update at the top by reading existing content

	// Get current text and find where logs start
	currentText := panelX.GetText(false)
	logsStart := 0

	// Clear and write new content
	panelX.Clear()

	// Restore existing logs
	if logsStart >= 0 && logsStart < len(currentText) {
		fmt.Fprint(panelX, currentText[logsStart:])
	}
}

func UpdateItemsPanel(panel *tview.TextView) {
	panel.Clear()

	gs := GetGameState()
	if gs.gameOver {
		fmt.Fprintf(panel, "[red]GAME OVER[white]\n\n")
	}

	fmt.Fprintf(panel, "[gray](s/w: move, u: upgrade)[white]\n\n")

	// Display items list with selection
	for i, itemName := range gs.itemsPanelItems {
		if i == gs.itemsPanelSelected {
			fmt.Fprintf(panel, "[black:white]%s[white:-]\n", itemName)
		} else {
			fmt.Fprintf(panel, "%s\n", itemName)
		}
	}
}

func UpdateShopPanel(panel *tview.TextView, selectedItem int, category int) {
	panel.Clear()

	categoryNames := []string{"COINS", "DIAMONDS", "GUNS"}
	items := GetAvailableItemsByCategory(category)

	// Show category tabs
	fmt.Fprintf(panel, "[gray](←/→: category, ↑/↓: item, I: buy)[white]\n\n")
	for i, name := range categoryNames {
		if i == category {
			fmt.Fprintf(panel, "[black:white]%s[white:-] ", name)
		} else {
			fmt.Fprintf(panel, "[gray]%s[white] ", name)
		}
	}
	fmt.Fprintf(panel, "\n\n")

	for i, item := range items {
		color := GetItemColor(item)

		// Build cost string
		costStr := ""
		if item.costCoins > 0 && item.costDiamonds > 0 {
			costStr = fmt.Sprintf("%dc+%dd", item.costCoins, item.costDiamonds)
		} else if item.costCoins > 0 {
			costStr = fmt.Sprintf("%dc", item.costCoins)
		} else if item.costDiamonds > 0 {
			costStr = fmt.Sprintf("%dd", item.costDiamonds)
		}

		// Build level string
		lvlStr := ""
		if item.maxLevel < 999 {
			lvlStr = fmt.Sprintf("%d/%d", item.currentLevel, item.maxLevel)
		} else {
			lvlStr = "-"
		}

		// Highlight selected item with background
		if i == selectedItem {
			fmt.Fprintf(panel, "[black:white]%s%s(%s/%s)[white:-]\n", color, item.name, costStr, lvlStr)
		} else {
			fmt.Fprintf(panel, "%s%s(%s/%s)[white]\n", color, item.name, costStr, lvlStr)
		}

		// Show description
		fmt.Fprintf(panel, "  %s\n", item.description)
	}
}

func UpdateRoomDefensePanel(panel *tview.TextView) {
	panel.Clear()

	gs := GetGameState()
	room := gs.rooms[gs.currentRoom]

	fmt.Fprintf(panel, "[yellow]DREAMERS[white]\n\n")

	// Show player first
	playerBar := DrawHPBar(gs.playerDefense, gs.playerMaxDefense, 15)
	fmt.Fprintf(panel, "[green]You[white] %s %d/%d\n", playerBar, gs.playerDefense, gs.playerMaxDefense)
	doorBar := DrawHPBar(gs.doorHP, gs.doorMaxHP, 15)
	fmt.Fprintf(panel, "Door Lv%d %s %d/%d\n\n", gs.doorLevel, doorBar, gs.doorHP, gs.doorMaxHP)

	// Show AI characters
	for _, char := range room.characters {
		fmt.Fprintf(panel, "[cyan]%-8s[white] Door Lv%d %d/%d\n", char.name, char.doorLevel, char.doorHP, char.doorMaxHP)
	}
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
		fmt.Fprintf(panel, "\n[red]⚠ HUNTER LEVEL %d[white]\n", gs.hunterLevel)
		fmt.Fprintf(panel, "Attack: %d every 3s\n", gs.hunterAttack)

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

// UpdateResourcePanel updates the top resource panel
func UpdateResourcePanel(panel *tview.TextView) {
	panel.Clear()

	gs := GetGameState()

	fmt.Fprintf(panel, "[gold]Coins: %d (+%.1f/s)[white]  [cyan]Diamonds: %d (+%.1f/s)[white]  [orange]Defense: %d[white]",
		gs.coins, gs.coinsPerS, gs.diamonds, gs.diamPerS, gs.playerMaxDefense)
}
