package main

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

// Game state
type GameState struct {
	coins     int
	diamonds  int
	coinsPerS float64
	diamPerS  float64

	// Items
	doorLevel    int
	doorHP       int
	doorMaxHP    int
	bedLevel     int
	playboxLevel int

	// Guns
	guns []Gun

	// Hunter
	hunterHP       int
	hunterMaxHP    int
	hunterPos      int
	hunterActive   bool
	hunterLevel    int
	hunterAttack   int
	lastAttackTime time.Time

	// Rooms (for spectate)
	currentRoom int
	rooms       []Room

	// Player defense (life)
	playerDefense    int
	playerMaxDefense int

	// Game state
	gameOver bool
	gameWon  bool

	// Your Items panel selection
	itemsPanelSelected int
	itemsPanelItems    []string
}

type Character struct {
	name            string
	defense         int
	maxDefense      int
	doorHP          int
	doorMaxHP       int
	doorLevel       int
	lastUpgradeTime time.Time
}

type Gun struct {
	name        string
	level       int
	damage      int
	attackSpeed float64 // attacks per second
	lastShot    time.Time
}

type Room struct {
	name       string
	items      []string
	characters []Character
	coinsPerS  float64
	diamPerS   float64
}

type Item struct {
	name         string
	currentLevel int
	maxLevel     int
	costCoins    int
	costDiamonds int
	production   float64
	description  string
	itemType     string  // "bed", "door", "playbox", "trap", "guard", "gun"
	damage       int     // for guns
	attackSpeed  float64 // for guns
}

var gameState *GameState

func InitGame() {
	gameState = &GameState{
		coins:              0,
		diamonds:           0,
		coinsPerS:          1,
		diamPerS:           0,
		doorLevel:          1,
		doorHP:             GetDoorHP(1),
		doorMaxHP:          GetDoorHP(1),
		bedLevel:           1,
		playboxLevel:       0,
		guns:               []Gun{},
		hunterHP:           0,
		hunterMaxHP:        0,
		hunterPos:          0,
		hunterActive:       false,
		hunterLevel:        1,
		hunterAttack:       GetHunterAttack(1),
		lastAttackTime:     time.Now(),
		currentRoom:        0,
		playerDefense:      100,
		playerMaxDefense:   100,
		gameOver:           false,
		gameWon:            false,
		itemsPanelSelected: 0,
		itemsPanelItems:    []string{},
		rooms: []Room{
			{
				name:  "Dream Realm",
				items: []string{},
				characters: []Character{
					{name: "Luna", defense: 80, maxDefense: 80, doorHP: GetDoorHP(1), doorMaxHP: GetDoorHP(1), doorLevel: 1, lastUpgradeTime: time.Now()},
					{name: "Morpheus", defense: 90, maxDefense: 90, doorHP: GetDoorHP(1), doorMaxHP: GetDoorHP(1), doorLevel: 1, lastUpgradeTime: time.Now()},
					{name: "Nyx", defense: 70, maxDefense: 70, doorHP: GetDoorHP(1), doorMaxHP: GetDoorHP(1), doorLevel: 1, lastUpgradeTime: time.Now()},
					{name: "Hypnos", defense: 85, maxDefense: 85, doorHP: GetDoorHP(1), doorMaxHP: GetDoorHP(1), doorLevel: 1, lastUpgradeTime: time.Now()},
				},
				coinsPerS: 0,
				diamPerS:  0,
			},
		},
	}
	updateItemsPanelList()
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
				gameState.gameOver = true
				gameState.gameWon = true
				AddLog(logPanel, "[green]Dream Hunter defeated! YOU WIN![white]")
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
		// Find dreamers with doors still standing
		aliveDreamers := []int{}
		for i := range gameState.rooms[0].characters {
			if gameState.rooms[0].characters[i].doorHP > 0 {
				aliveDreamers = append(aliveDreamers, i)
			}
		}

		// Attack one random dreamer
		if len(aliveDreamers) > 0 {
			targetIdx := aliveDreamers[now.Unix()%int64(len(aliveDreamers))]
			char := &gameState.rooms[0].characters[targetIdx]

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

			// Dreamers upgrade their doors every 30 seconds
			if char.doorLevel < 10 && now.Sub(char.lastUpgradeTime) >= 30*time.Second {
				char.doorLevel++
				char.doorMaxHP = GetDoorHP(char.doorLevel)
				char.doorHP = char.doorMaxHP
				char.lastUpgradeTime = now
			}
		}
	}
}

func SpawnHunter(logPanel *tview.TextView) {
	if !gameState.hunterActive && !gameState.gameOver {
		gameState.hunterActive = true
		gameState.hunterHP = GetHunterHP(gameState.hunterLevel)
		gameState.hunterMaxHP = gameState.hunterHP
		gameState.hunterAttack = GetHunterAttack(gameState.hunterLevel)
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

// GetHunterHP calculates hunter HP based on level
// Formula: HP(L) = HP₀ × r^(L−1), HP₀=500, r=1.4
func GetHunterHP(level int) int {
	hp0 := 500.0
	r := 1.4
	hp := hp0 * pow(r, float64(level-1))
	return int(hp)
}

// GetHunterAttack calculates hunter attack based on level
// Formula: ATK(L) = ATK₀ × s^(L−1), ATK₀=50, s=1.25
func GetHunterAttack(level int) int {
	atk0 := 50.0
	s := 1.25
	atk := atk0 * pow(s, float64(level-1))
	return int(atk)
}

// GetDoorHP calculates door HP based on level
// Formula: HP(L) = HP₀ + a × (L−1), HP₀=2000, a=300
func GetDoorHP(level int) int {
	hp0 := 2000
	a := 300
	return hp0 + a*(level-1)
}

// GetGunDamage calculates gun damage based on level (number of guns owned)
// Formula: GunDamage(L) = G₀ × t^(L−1), G₀=30, t=1.2
func GetGunDamage(gunLevel int) int {
	g0 := 30.0
	t := 1.2
	if gunLevel == 0 {
		gunLevel = 1
	}
	damage := g0 * pow(t, float64(gunLevel-1))
	return int(damage)
}

// pow is a simple power function for float64
func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := 1.0
	absExp := exp
	if exp < 0 {
		absExp = -exp
	}
	for i := 0; i < int(absExp); i++ {
		result *= base
	}
	if exp < 0 {
		return 1 / result
	}
	return result
}

// GetGunPrice calculates gun price based on level
// Formula: 8 → 16 (×2), then 16 → 40 (+24), 40 → 88 (+48), 88 → 176 (+88), etc.
// The increment itself doubles each time: +24, +48, +96, +192, +384...
func GetGunPrice(gunCount int) int {
	if gunCount == 0 {
		return 8
	}

	price := 8
	increment := 8 // First increment after base price (8 → 16 = +8, which is base price doubled = 16)

	for i := 0; i < gunCount; i++ {
		if i == 0 {
			// First upgrade: 8 → 16 (×2)
			price = 16
			increment = 24 // Next increment
		} else {
			// Subsequent upgrades follow pattern where increment grows
			price += increment
			increment = price - 16 // New increment is current price minus the second price
		}
	}

	return price
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
		gameState.doorMaxHP = GetDoorHP(gameState.doorLevel)
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
			name:        item.name,
			level:       1,
			damage:      item.damage,
			attackSpeed: item.attackSpeed,
			lastShot:    time.Now(),
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
		// Bed: levels 1-10, price = 25 * 2^(current_level-1)
		if gameState.bedLevel < 10 {
			nextLevel := gameState.bedLevel + 1
			costShift := uint(gameState.bedLevel - 1)
			coinCost := 25 * (int(1) << costShift)
			diamondCost := 0
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

		// Door: levels 1-10, price = 16 * 2^(current_level-1)
		if gameState.doorLevel < 10 {
			costShift := uint(gameState.doorLevel - 1)
			coinCost := 16 * (int(1) << costShift)

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
		// Playbox: levels 1-10, price = 200 * 2^(n-1)
		if gameState.playboxLevel < 10 {
			nextLevel := gameState.playboxLevel + 1
			costShift := uint(nextLevel - 1)
			coinCost := 200 * (int(1) << costShift)
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
		gunCount := len(gameState.guns)
		gunPrice := GetGunPrice(gunCount)
		gunDamage := GetGunDamage(gunCount + 1)

		items = append(items, Item{
			name:         "Pistol",
			currentLevel: gunCount,
			maxLevel:     999,
			costCoins:    gunPrice,
			costDiamonds: 0,
			damage:       gunDamage,
			attackSpeed:  1.0,
			description:  fmt.Sprintf("%ddmg 1.0atk/s", gunDamage),
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
		gameState.doorMaxHP = GetDoorHP(gameState.doorLevel)
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
			name:        item.name,
			level:       1,
			damage:      item.damage,
			attackSpeed: item.attackSpeed,
			lastShot:    time.Now(),
		}
		gameState.guns = append(gameState.guns, gun)
		AddLog(logPanel, fmt.Sprintf("[yellow]%s purchased! Damage: %d, Speed: %.1f/s[white]", item.name, item.damage, item.attackSpeed))
	}

	updateItemsPanelList()
}

func GetAvailableItems() []Item {
	items := []Item{}

	// Bed: levels 1-10, price = 25 * 2^(current_level-1)
	// Production: 1, 2, 4, 8, 16, 32, 64, 128, 256, 512 coins/s
	if gameState.bedLevel < 10 {
		nextLevel := gameState.bedLevel + 1
		costShift := uint(gameState.bedLevel - 1)
		coinCost := 25 * (int(1) << costShift)
		diamondCost := 0
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

	// Door: levels 1-10, price = 16 * 2^(current_level-1)
	if gameState.doorLevel < 10 {
		nextLevel := gameState.doorLevel + 1
		costShift := uint(gameState.doorLevel - 1)
		coinCost := 16 * (int(1) << costShift)

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

	// Playbox: levels 1-10, price = 200 * 2^(n-1)
	// Production: 1, 2, 4, 8, 16, 32, 64, 128, 256, 512 diamonds/s
	if gameState.playboxLevel < 10 {
		nextLevel := gameState.playboxLevel + 1
		costShift := uint(nextLevel - 1)
		coinCost := 200 * (int(1) << costShift)
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
	gunCount := len(gameState.guns)
	gunPrice := GetGunPrice(gunCount)
	gunDamage := GetGunDamage(gunCount + 1)

	items = append(items, Item{
		name:         "Pistol",
		currentLevel: gunCount,
		maxLevel:     999,
		costCoins:    gunPrice,
		costDiamonds: 0,
		damage:       gunDamage,
		attackSpeed:  1.0,
		description:  fmt.Sprintf("%d dmg, 1.0 atk/s", gunDamage),
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

// updateItemsPanelList updates the items panel list
func updateItemsPanelList() {
	items := []string{}

	// Add door
	items = append(items, fmt.Sprintf("Door Lv%d (HP:%d)", gameState.doorLevel, gameState.doorMaxHP))

	// Add bed if purchased
	if gameState.bedLevel > 0 {
		items = append(items, fmt.Sprintf("Bed Lv%d (+%.0f/s)", gameState.bedLevel, gameState.coinsPerS))
	}

	// Add playbox if purchased
	if gameState.playboxLevel > 0 {
		items = append(items, fmt.Sprintf("Playbox Lv%d (+%.0f/s)", gameState.playboxLevel, gameState.diamPerS))
	}

	// Add defense
	items = append(items, fmt.Sprintf("Defense: %d", gameState.playerMaxDefense))

	// Add guns
	for _, gun := range gameState.guns {
		items = append(items, fmt.Sprintf("%s (D:%d S:%.1f)", gun.name, gun.damage, gun.attackSpeed))
	}

	gameState.itemsPanelItems = items
}

// MoveItemSelection moves the selection in items panel
func MoveItemSelection(direction int) {
	newPos := gameState.itemsPanelSelected + direction
	if newPos >= 0 && newPos < len(gameState.itemsPanelItems) {
		gameState.itemsPanelSelected = newPos
	}
}

// UpgradeSelectedItem upgrades the selected item in items panel
func UpgradeSelectedItem(logPanel *tview.TextView) {
	if gameState.itemsPanelSelected < 0 || gameState.itemsPanelSelected >= len(gameState.itemsPanelItems) {
		return
	}

	// Determine which item to upgrade based on selection
	itemOffset := 0

	// Item 0 is always door
	if gameState.itemsPanelSelected == itemOffset {
		// Door
		if gameState.doorLevel < 10 {
			costShift := uint(gameState.doorLevel - 1)
			coinCost := 16 * (int(1) << costShift)

			if gameState.coins >= coinCost {
				gameState.coins -= coinCost
				gameState.doorLevel++
				gameState.doorMaxHP = GetDoorHP(gameState.doorLevel)
				gameState.doorHP = gameState.doorMaxHP
				AddLog(logPanel, fmt.Sprintf("[green]Door upgraded to level %d! (HP: %d)[white]", gameState.doorLevel, gameState.doorMaxHP))
				updateItemsPanelList()
			} else {
				AddLog(logPanel, "[red]Not enough coins![white]")
			}
		} else {
			AddLog(logPanel, "[yellow]Door is at max level![white]")
		}
		return
	}
	itemOffset++

	// Bed (if exists)
	if gameState.bedLevel > 0 {
		if gameState.itemsPanelSelected == itemOffset {
			if gameState.bedLevel < 10 {
				costShift := uint(gameState.bedLevel - 1)
				coinCost := 25 * (int(1) << costShift)

				if gameState.coins >= coinCost {
					gameState.coins -= coinCost
					gameState.bedLevel++
					AddLog(logPanel, fmt.Sprintf("[green]Bed upgraded to level %d![white]", gameState.bedLevel))
					updateItemsPanelList()
				} else {
					AddLog(logPanel, "[red]Not enough coins![white]")
				}
			} else {
				AddLog(logPanel, "[yellow]Bed is at max level![white]")
			}
			return
		}
		itemOffset++
	}

	// Playbox (if exists)
	if gameState.playboxLevel > 0 {
		if gameState.itemsPanelSelected == itemOffset {
			if gameState.playboxLevel < 10 {
				nextLevel := gameState.playboxLevel + 1
				costShift := uint(nextLevel - 1)
				coinCost := 200 * (int(1) << costShift)

				if gameState.coins >= coinCost {
					gameState.coins -= coinCost
					gameState.playboxLevel++
					AddLog(logPanel, fmt.Sprintf("[green]Playbox upgraded to level %d![white]", gameState.playboxLevel))
					updateItemsPanelList()
				} else {
					AddLog(logPanel, "[red]Not enough coins![white]")
				}
			} else {
				AddLog(logPanel, "[yellow]Playbox is at max level![white]")
			}
			return
		}
		itemOffset++
	}
}
