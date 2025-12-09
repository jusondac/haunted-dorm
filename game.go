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
	
	// Guns
	guns       []Gun
	
	// Hunter
	hunterHP   int
	hunterMaxHP int
	hunterPos  int
	hunterActive bool
	hunterLevel int
	hunterAttack int
	lastAttackTime time.Time
	
	// Rooms (for spectate)
	currentRoom int
	rooms      []Room
	
	// Player defense (life)
	playerDefense int
	playerMaxDefense int
	
	// Game state
	gameOver bool
	gameWon  bool
}

type Character struct {
	name      string
	defense   int
	maxDefense int
	doorHP    int
	doorMaxHP int
}

type Gun struct {
	name       string
	level      int
	damage     int
	attackSpeed float64 // attacks per second
	lastShot   time.Time
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
	itemType    string // "bed", "door", "playbox", "trap", "guard", "gun"
	damage      int    // for guns
	attackSpeed float64 // for guns
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
		guns:       []Gun{},
		hunterHP:   0,
		hunterMaxHP: 0,
		hunterPos:  0,
		hunterActive: false,
		hunterLevel: 1,
		hunterAttack: 10,
		lastAttackTime: time.Now(),
		currentRoom: 0,
		playerDefense: 100,
		playerMaxDefense: 100,
		gameOver: false,
		gameWon: false,
		rooms: []Room{
			{
				name: "Dream Realm",
				items: []string{},
				characters: []Character{
					{name: "Luna", defense: 80, maxDefense: 80, doorHP: 100, doorMaxHP: 100},
					{name: "Morpheus", defense: 90, maxDefense: 90, doorHP: 120, doorMaxHP: 120},
					{name: "Nyx", defense: 70, maxDefense: 70, doorHP: 90, doorMaxHP: 90},
					{name: "Hypnos", defense: 85, maxDefense: 85, doorHP: 110, doorMaxHP: 110},
				},
				coinsPerS: 0,
				diamPerS: 0,
			},
		},
	}
}

func UpdateGame() {
	if gameState.gameOver {
		return
	}
	
	// Calculate coins per second from beds
	gameState.coinsPerS = 0
	if gameState.bedLevel > 0 {
		shift := uint(gameState.bedLevel - 1)
		gameState.coinsPerS = float64(int(1) << shift) // 2^(level-1): 1,2,4,8,16...
	}
	
	// Calculate diamonds per second from playbox
	gameState.diamPerS = 0
	if gameState.playboxLevel > 0 {
		shift := uint(gameState.playboxLevel - 1)
		gameState.diamPerS = float64(int(1) << shift) // 2^(level-1): 1,2,4,8,16...
	}
	
	// Add coins
	gameState.coins += int(gameState.coinsPerS)
	gameState.diamonds += int(gameState.diamPerS)
}

func UpdateCombat(logPanel *tview.TextView) {
	if !gameState.hunterActive || gameState.gameOver {
		return
	}
	
	now := time.Now()
	
	// Guns shoot at hunter
	for i := range gameState.guns {
		gun := &gameState.guns[i]
		interval := time.Duration(1000.0/gun.attackSpeed) * time.Millisecond
		
		if now.Sub(gun.lastShot) >= interval {
			gameState.hunterHP -= gun.damage
			gun.lastShot = now
			
			if gameState.hunterHP <= 0 {
				gameState.hunterHP = 0
				gameState.hunterActive = false
				AddLog(logPanel, "[green]Dream Hunter defeated![white]")
				
				// Hunter upgrades after defeat
				gameState.hunterLevel++
				gameState.hunterMaxHP += 20
				gameState.hunterAttack += 5
				AddLog(logPanel, fmt.Sprintf("[yellow]Hunter upgraded to level %d![white]", gameState.hunterLevel))
				return
			}
		}
	}
	
	// Hunter attacks door every 3 seconds
	if now.Sub(gameState.lastAttackTime) >= 3*time.Second {
		gameState.doorHP -= gameState.hunterAttack
		gameState.lastAttackTime = now
		AddLog(logPanel, fmt.Sprintf("[red]Hunter attacks door! -%d HP[white]", gameState.hunterAttack))
		
		if gameState.doorHP <= 0 {
			gameState.doorHP = 0
			gameState.gameOver = true
			AddLog(logPanel, "[red]GAME OVER! Your door is broken![white]")
			return
		}
	}
	
	// Hunter attacks other dreamers' doors randomly
	if now.Unix()%5 == 0 { // Every 5 seconds
		for i := range gameState.rooms[0].characters {
			char := &gameState.rooms[0].characters[i]
			if char.doorHP > 0 {
				damage := gameState.hunterAttack / 2
				char.doorHP -= damage
				if char.doorHP < 0 {
					char.doorHP = 0
				}
				
				// Dreamers repair their doors slowly
				if char.doorHP < char.doorMaxHP && char.doorHP > 0 {
					char.doorHP += 2
					if char.doorHP > char.doorMaxHP {
						char.doorHP = char.doorMaxHP
					}
				}
			}
		}
	}
}

func SpawnHunter(logPanel *tview.TextView) {
	if !gameState.hunterActive && !gameState.gameOver {
		gameState.hunterActive = true
		gameState.hunterHP = 50 + (gameState.hunterLevel * 20)
		gameState.hunterMaxHP = gameState.hunterHP
		gameState.hunterPos = 0
		gameState.lastAttackTime = time.Now()
		AddLog(logPanel, fmt.Sprintf("[red]Dream Hunter Level %d spawned![white]", gameState.hunterLevel))
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
	case "gun":
		gun := Gun{
			name:       item.name,
			level:      1,
			damage:     item.damage,
			attackSpeed: item.attackSpeed,
			lastShot:   time.Now(),
		}
		gameState.guns = append(gameState.guns, gun)
		AddLog(logPanel, fmt.Sprintf("[yellow]%s purchased! Damage: %d, Speed: %.1f/s[white]", item.name, item.damage, item.attackSpeed))
	}
}

func GetGameState() *GameState {
	return gameState
}

func GetAvailableItemsByCategory(category int) []Item {
	items := []Item{}
	
	switch category {
	case 0: // Coins category
		// Bed: levels 1-10, price starts at 20c and doubles, costs diamonds from level 3
		if gameState.bedLevel < 10 {
			nextLevel := gameState.bedLevel + 1
			baseCost := 20 * (1 << uint(gameState.bedLevel))
			coinCost := baseCost
			diamondCost := 0
			if nextLevel >= 3 {
				diamondShift := uint(nextLevel - 3)
				diamondCost = 10 * (int(1) << diamondShift)
			}
			prodShift := uint(nextLevel - 1)
			production := float64(int(1) << prodShift)
			
			items = append(items, Item{
				name:         "Bed",
				currentLevel: gameState.bedLevel,
				maxLevel:     10,
				costCoins:    coinCost,
				costDiamonds: diamondCost,
				production:   production,
				description:  fmt.Sprintf("+%.0f coins/s", production),
				itemType:     "bed",
			})
		}
		
		// Door: levels 1-10
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
				description:  "+50 HP",
				itemType:     "door",
			})
		}
		
	case 1: // Diamonds category
		// Playbox: levels 1-10
		if gameState.playboxLevel < 10 {
			nextLevel := gameState.playboxLevel + 1
			costShift := uint(gameState.playboxLevel)
			coinCost := 10 * (int(1) << costShift)
			prodShift := uint(nextLevel - 1)
			production := float64(int(1) << prodShift)
			
			items = append(items, Item{
				name:         "Playbox",
				currentLevel: gameState.playboxLevel,
				maxLevel:     10,
				costCoins:    coinCost,
				costDiamonds: 0,
				production:   production,
				description:  fmt.Sprintf("+%.0f diamonds/s", production),
				itemType:     "playbox",
			})
		}
		
		// Trap
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
		
		// Guard
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
		
	case 2: // Guns category
		items = append(items, Item{
			name:         "Pistol",
			currentLevel: 0,
			maxLevel:     999,
			costCoins:    50,
			costDiamonds: 0,
			damage:       5,
			attackSpeed:  1.0,
			description:  "5dmg 1.0atk/s",
			itemType:     "gun",
		})
		
		items = append(items, Item{
			name:         "Rifle",
			currentLevel: 0,
			maxLevel:     999,
			costCoins:    150,
			costDiamonds: 5,
			damage:       15,
			attackSpeed:  0.5,
			description:  "15dmg 0.5atk/s",
			itemType:     "gun",
		})
		
		items = append(items, Item{
			name:         "Shotgun",
			currentLevel: 0,
			maxLevel:     999,
			costCoins:    200,
			costDiamonds: 10,
			damage:       30,
			attackSpeed:  0.3,
			description:  "30dmg 0.3atk/s",
			itemType:     "gun",
		})
		
		items = append(items, Item{
			name:         "Machine Gun",
			currentLevel: 0,
			maxLevel:     999,
			costCoins:    300,
			costDiamonds: 20,
			damage:       8,
			attackSpeed:  3.0,
			description:  "8dmg 3.0atk/s",
			itemType:     "gun",
		})
		
		items = append(items, Item{
			name:         "Sniper",
			currentLevel: 0,
			maxLevel:     999,
			costCoins:    500,
			costDiamonds: 50,
			damage:       100,
			attackSpeed:  0.2,
			description:  "100dmg 0.2atk/s",
			itemType:     "gun",
		})
	}
	
	return items
}

func BuyItemByCategory(itemIndex int, category int, logPanel *tview.TextView) {
	items := GetAvailableItemsByCategory(category)
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
	case "gun":
		gun := Gun{
			name:       item.name,
			level:      1,
			damage:     item.damage,
			attackSpeed: item.attackSpeed,
			lastShot:   time.Now(),
		}
		gameState.guns = append(gameState.guns, gun)
		AddLog(logPanel, fmt.Sprintf("[yellow]%s purchased! Damage: %d, Speed: %.1f/s[white]", item.name, item.damage, item.attackSpeed))
	}
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
			diamondShift := uint(nextLevel - 3)
			diamondCost = 10 * (int(1) << diamondShift) // 0,0,10,20,40,80,160,320,640,1280
		}
		prodShift := uint(nextLevel - 1)
		production := float64(int(1) << prodShift)
		
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
		costShift := uint(gameState.playboxLevel)
		coinCost := 10 * (int(1) << costShift) // 10, 20, 40, 80, 160...
		prodShift := uint(nextLevel - 1)
		production := float64(int(1) << prodShift)
		
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
	
	// Guns - various weapons
	items = append(items, Item{
		name:         "Pistol",
		currentLevel: len(gameState.guns),
		maxLevel:     999,
		costCoins:    50,
		costDiamonds: 0,
		damage:       5,
		attackSpeed:  1.0,
		description:  "5 dmg, 1.0 atk/s",
		itemType:     "gun",
	})
	
	items = append(items, Item{
		name:         "Rifle",
		currentLevel: 0,
		maxLevel:     999,
		costCoins:    150,
		costDiamonds: 5,
		damage:       15,
		attackSpeed:  0.5,
		description:  "15 dmg, 0.5 atk/s",
		itemType:     "gun",
	})
	
	items = append(items, Item{
		name:         "Shotgun",
		currentLevel: 0,
		maxLevel:     999,
		costCoins:    200,
		costDiamonds: 10,
		damage:       30,
		attackSpeed:  0.3,
		description:  "30 dmg, 0.3 atk/s",
		itemType:     "gun",
	})
	
	items = append(items, Item{
		name:         "Machine Gun",
		currentLevel: 0,
		maxLevel:     999,
		costCoins:    300,
		costDiamonds: 20,
		damage:       8,
		attackSpeed:  3.0,
		description:  "8 dmg, 3.0 atk/s",
		itemType:     "gun",
	})
	
	items = append(items, Item{
		name:         "Sniper",
		currentLevel: 0,
		maxLevel:     999,
		costCoins:    500,
		costDiamonds: 50,
		damage:       100,
		attackSpeed:  0.2,
		description:  "100 dmg, 0.2 atk/s",
		itemType:     "gun",
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
