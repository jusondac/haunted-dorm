package main

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

// Game state
type GameState struct {
	coins      int
	diamonds   int
	coinsPerS  float64
	diamPerS   float64
	
	// Items
	doorLevel  int
	doorHP     int
	doorMaxHP  int
	bedLevel   int
	beds       int
	
	// Hunter
	hunterHP   int
	hunterMaxHP int
	hunterPos  int
	hunterActive bool
	
	// Rooms (for spectate)
	currentRoom int
	rooms      []Room
}

type Room struct {
	name       string
	items      []string
	defense    int
	coinsPerS  float64
	diamPerS   float64
}

type Item struct {
	name       string
	level      int
	cost       int
	costType   string // "coin" or "diamond"
	description string
}

var gameState *GameState
var availableItems []Item

func InitGame() {
	gameState = &GameState{
		coins:      100,
		diamonds:   0,
		coinsPerS:  0,
		diamPerS:   0,
		doorLevel:  1,
		doorHP:     100,
		doorMaxHP:  100,
		bedLevel:   0,
		beds:       0,
		hunterHP:   0,
		hunterMaxHP: 0,
		hunterPos:  0,
		hunterActive: false,
		currentRoom: 0,
		rooms: []Room{
			{name: "Main Room", items: []string{"Door Lv1"}, defense: 10, coinsPerS: 0, diamPerS: 0},
			{name: "Bedroom", items: []string{}, defense: 5, coinsPerS: 0, diamPerS: 0},
			{name: "Kitchen", items: []string{}, defense: 3, coinsPerS: 0, diamPerS: 0},
		},
	}
	
	availableItems = []Item{
		{name: "Bed", level: 1, cost: 50, costType: "coin", description: "+1 coin/s"},
		{name: "Door Upgrade", level: 2, cost: 100, costType: "coin", description: "Upgrade door HP"},
		{name: "Trap", level: 1, cost: 5, costType: "diamond", description: "+5 defense"},
		{name: "Guard", level: 1, cost: 10, costType: "diamond", description: "+10 defense"},
	}
}

func UpdateGame() {
	// Update coins per second
	gameState.coinsPerS = float64(gameState.beds)
	
	// Add coins
	gameState.coins += int(gameState.coinsPerS)
	gameState.diamonds += int(gameState.diamPerS)
	
	// Update hunter
	if gameState.hunterActive {
		gameState.hunterPos++
		if gameState.hunterPos >= 10 {
			// Hunter reaches door
			gameState.doorHP -= 10
			if gameState.doorHP <= 0 {
				gameState.doorHP = 0
				// Game over logic
			}
			gameState.hunterPos = 0
		}
	}
}

func SpawnHunter(logPanel *tview.TextView) {
	if !gameState.hunterActive {
		gameState.hunterActive = true
		gameState.hunterHP = 50 + (gameState.doorLevel * 10)
		gameState.hunterMaxHP = gameState.hunterHP
		gameState.hunterPos = 0
		AddLog(logPanel, "[red]Dream Hunter spawned![white]")
	}
}

func AddLog(logPanel *tview.TextView, message string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(logPanel, "[yellow]%s[white] %s\n", timestamp, message)
	logPanel.ScrollToEnd()
}

func DrawHPBar(current, max int, width int) string {
	if max == 0 {
		return ""
	}
	filled := int(float64(current) / float64(max) * float64(width))
	if filled < 0 {
		filled = 0
	}
	bar := "["
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"
	return bar
}

func BuyItem(itemIndex int, logPanel *tview.TextView) {
	if itemIndex < 0 || itemIndex >= len(availableItems) {
		AddLog(logPanel, "[red]Invalid item![white]")
		return
	}
	
	item := availableItems[itemIndex]
	
	// Check if can afford
	canAfford := false
	if item.costType == "coin" && gameState.coins >= item.cost {
		canAfford = true
		gameState.coins -= item.cost
	} else if item.costType == "diamond" && gameState.diamonds >= item.cost {
		canAfford = true
		gameState.diamonds -= item.cost
	}
	
	if !canAfford {
		AddLog(logPanel, fmt.Sprintf("[red]Not enough %ss![white]", item.costType))
		return
	}
	
	// Apply item effect
	switch item.name {
	case "Bed":
		gameState.beds++
		gameState.bedLevel++
		AddLog(logPanel, fmt.Sprintf("[green]Bought Bed! Total: %d[white]", gameState.beds))
	case "Door Upgrade":
		gameState.doorLevel++
		gameState.doorMaxHP += 50
		gameState.doorHP = gameState.doorMaxHP
		AddLog(logPanel, fmt.Sprintf("[green]Door upgraded to level %d![white]", gameState.doorLevel))
	case "Trap":
		gameState.rooms[0].defense += 5
		AddLog(logPanel, "[green]Trap installed![white]")
	case "Guard":
		gameState.rooms[0].defense += 10
		AddLog(logPanel, "[green]Guard hired![white]")
	}
}

func GetGameState() *GameState {
	return gameState
}

func GetAvailableItems() []Item {
	return availableItems
}
