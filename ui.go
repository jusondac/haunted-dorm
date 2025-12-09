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

	fmt.Fprintf(panel, "[yellow]YOUR ITEMS[white]\n\n")

	if gs.gameOver {
		fmt.Fprintf(panel, "[red]GAME OVER[white]\n\n")
	}

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

	// Guns
	if len(gs.guns) > 0 {
		fmt.Fprintf(panel, "\n[yellow]GUNS:[white]\n")
		for _, gun := range gs.guns {
			fmt.Fprintf(panel, "• %s (Dmg:%d Spd:%.1f)\n", gun.name, gun.damage, gun.attackSpeed)
		}
	}
}

func UpdateShopPanel(panel *tview.TextView, selectedItem int, category int) {
	panel.Clear()

	categoryNames := []string{"COINS", "DIAMONDS", "GUNS"}
	items := GetAvailableItemsByCategory(category)

	// Show category tabs
	fmt.Fprintf(panel, "[yellow]SHOP[white] [gray](←/→: category, ↑/↓: item, I: buy)[white]\n")
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
		charBar := DrawHPBar(char.defense, char.maxDefense, 15)
		fmt.Fprintf(panel, "[cyan]%-8s[white] %s %d/%d\n", char.name, charBar, char.defense, char.maxDefense)
		charDoorBar := DrawHPBar(char.doorHP, char.doorMaxHP, 15)
		fmt.Fprintf(panel, "Door Lv%d %s %d/%d\n", char.doorLevel, charDoorBar, char.doorHP, char.doorMaxHP)
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
