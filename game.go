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
	playboxLevel int
	
	// Hunter
	hunterHP   int
	hunterMaxHP int
	hunterPos  int
	hunterActive bool
	
	// Rooms (for spectate)
	currentRoom int
	rooms      []Room
	
	// Player defense (life)
	playerDefense int
	playerMaxDefense int
}

type Character struct {
	name      string
	defense   int
	maxDefense int
}

type Room struct {
	name       string
	items      []string
	characters []Character
	coinsPerS  float64
	diamPerS   float64
}

type Item struct {
	name        string
	currentLevel int
	maxLevel    int
	costCoins   int
	costDiamonds int
	production  float64
	description string
	itemType    string // "bed", "door", "playbox", "trap", "guard"
}

var gameState *GameState

func InitGame() {
	gameState = &GameState{
		coins:      100,
		diamonds:   0,
		coinsPerS:  0,
		diamPerS:   0,
		doorLevel:  0,
		doorHP:     100,
		doorMaxHP:  100,
		bedLevel:   0,
		playboxLevel: 0,
		hunterHP:   0,
		hunterMaxHP: 0,
		hunterPos:  0,
		hunterActive: false,
		currentRoom: 0,
		playerDefense: 100,
		playerMaxDefense: 100,
		rooms: []Room{
			{
				name: "Dream Realm",
				items: []string{},
				characters: []Character{
					{name: "Luna", defense: 80, maxDefense: 80},
					{name: "Morpheus", defense: 90, maxDefense: 90},
					{name: "Nyx", defense: 70, maxDefense: 70},
					{name: "Hypnos", defense: 85, maxDefense: 85},
				},
				coinsPerS: 0,
				diamPerS: 0,
			},
		},
	}
}

func UpdateGame() {
	// Calculate coins per second from beds
	gameState.coinsPerS = 0
	if gameState.bedLevel > 0 {
		gameState.coinsPerS = float64(1 << (gameState.bedLevel - 1)) // 2^(level-1): 1,2,4,8,16...
	}
	
	// Calculate diamonds per second from playbox
	gameState.diamPerS = 0
	if gameState.playboxLevel > 0 {
		gameState.diamPerS = float64(1 << (gameState.playboxLevel - 1)) // 2^(level-1): 1,2,4,8,16...
	}
	
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
	items := GetAvailableItems()
	if itemIndex < 0 || itemIndex >= len(items) {
		AddLog(logPanel, "[red]Invalid item![white]")
		return
	}
	
	item := items[itemIndex]
	
	// Check if can afford
	if !CanAffordItem(item) {
		AddLog(logPanel, "[red]Not enough resources![white]")
		return
	}
	
	// Deduct costs
	gameState.coins -= item.costCoins
	gameState.diamonds -= item.costDiamonds
	
	// Apply item effect
	switch item.itemType {
	case "bed":
		gameState.bedLevel++
		AddLog(logPanel, fmt.Sprintf("[green]Bed upgraded to level %d! (+%.0f coins/s)[white]", gameState.bedLevel, item.production))
	case "door":
		gameState.doorLevel++
		gameState.doorMaxHP += 50
		gameState.doorHP = gameState.doorMaxHP
		AddLog(logPanel, fmt.Sprintf("[green]Door upgraded to level %d! (HP: %d)[white]", gameState.doorLevel, gameState.doorMaxHP))
	case "playbox":
		gameState.playboxLevel++
		AddLog(logPanel, fmt.Sprintf("[cyan]Playbox upgraded to level %d! (+%.0f diamonds/s)[white]", gameState.playboxLevel, item.production))
	case "trap":
		gameState.playerDefense += 5
		gameState.playerMaxDefense += 5
		AddLog(logPanel, "[green]Trap installed! Defense +5[white]")
	case "guard":
		gameState.playerDefense += 10
		gameState.playerMaxDefense += 10
		AddLog(logPanel, "[green]Guard hired! Defense +10[white]")
	}
}

func GetGameState() *GameState {
	return gameState
}

func GetAvailableItems() []Item {
	items := []Item{}
	
	// Bed: levels 1-10, price starts at 20c and doubles, costs diamonds from level 3
	// Production: 1, 2, 4, 8, 16, 32, 64, 128, 256, 512 coins/s
	if gameState.bedLevel < 10 {
		nextLevel := gameState.bedLevel + 1
		baseCost := 20 * (1 << uint(gameState.bedLevel)) // 20, 40, 80, 160, 320...
		coinCost := baseCost
		diamondCost := 0
		if nextLevel >= 3 {
			diamondCost = 10 * (1 << uint(nextLevel - 3)) // 0,0,10,20,40,80,160,320,640,1280
		}
		production := float64(1 << uint(nextLevel - 1))
		
		items = append(items, Item{
			name:         "Bed",
			currentLevel: gameState.bedLevel,
			maxLevel:     10,
			costCoins:    coinCost,
			costDiamonds: diamondCost,
			production:   production,
			description:  fmt.Sprintf("Lv%d→%d: +%.0f coins/s", gameState.bedLevel, nextLevel, production),
			itemType:     "bed",
		})
	}
	
	// Door: levels 1-10, price starts at 10c and increases by 10c each level
	if gameState.doorLevel < 10 {
		nextLevel := gameState.doorLevel + 1
		coinCost := 10 * nextLevel
		
		items = append(items, Item{
			name:         "Door",
			currentLevel: gameState.doorLevel,
			maxLevel:     10,
			costCoins:    coinCost,
			costDiamonds: 0,
			production:   0,
			description:  fmt.Sprintf("Lv%d→%d: +50 HP", gameState.doorLevel, nextLevel),
			itemType:     "door",
		})
	}
	
	// Playbox: levels 1-10, price starts at 10c and doubles
	// Production: 1, 2, 4, 8, 16, 32, 64, 128, 256, 512 diamonds/s
	if gameState.playboxLevel < 10 {
		nextLevel := gameState.playboxLevel + 1
		coinCost := 10 * (1 << uint(gameState.playboxLevel)) // 10, 20, 40, 80, 160...
		production := float64(1 << uint(nextLevel - 1))
		
		items = append(items, Item{
			name:         "Playbox",
			currentLevel: gameState.playboxLevel,
			maxLevel:     10,
			costCoins:    coinCost,
			costDiamonds: 0,
			production:   production,
			description:  fmt.Sprintf("Lv%d→%d: +%.0f diamonds/s", gameState.playboxLevel, nextLevel, production),
			itemType:     "playbox",
		})
	}
	
	// Trap and Guard - one-time purchases
	items = append(items, Item{
		name:         "Trap",
		currentLevel: 0,
		maxLevel:     999,
		costCoins:    0,
		costDiamonds: 5,
		production:   0,
		description:  "+5 defense",
		itemType:     "trap",
	})
	
	items = append(items, Item{
		name:         "Guard",
		currentLevel: 0,
		maxLevel:     999,
		costCoins:    0,
		costDiamonds: 10,
		production:   0,
		description:  "+10 defense",
		itemType:     "guard",
	})
	
	return items
}

func CanAffordItem(item Item) bool {
	hasCoins := gameState.coins >= item.costCoins
	hasDiamonds := gameState.diamonds >= item.costDiamonds
	return hasCoins && hasDiamonds
}

func GetItemColor(item Item) string {
	// Check if owned (at max level for upgradeable items)
	if item.currentLevel >= item.maxLevel && item.maxLevel < 999 {
		return "[blue]"
	}
	
	// Check if can afford (upgradeable)
	if CanAffordItem(item) {
		return "[green]"
	}
	
	// Too expensive
	return "[red]"
}
