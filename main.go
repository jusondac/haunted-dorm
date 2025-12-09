package main

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()
	InitGame()

	// Left side panels
	panelLog := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})
	panelLog.SetBorder(true).SetTitle(" Status & Logs ").SetTitleAlign(tview.AlignLeft)
	
	panelRoomDefense := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	panelRoomDefense.SetBorder(true).SetTitle(" Dreamers ").SetTitleAlign(tview.AlignLeft)
	
	panelRoomItems := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	panelRoomItems.SetBorder(true).SetTitle(" Room Items ").SetTitleAlign(tview.AlignLeft)

	// Right side panels
	panelYourItems := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	panelYourItems.SetBorder(true).SetTitle(" Your Items ").SetTitleAlign(tview.AlignLeft)

	panelShop := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	panelShop.SetBorder(true).SetTitle(" Shop ").SetTitleAlign(tview.AlignLeft)

	selectedItem := 0
	shopCategory := 0 // 0=Coins, 1=Diamonds, 2=Guns
	
	// Function to update all panels
	updatePanels := func() {
		UpdateLogPanel(panelLog)
		UpdateItemsPanel(panelYourItems)
		UpdateShopPanel(panelShop, selectedItem, shopCategory)
		UpdateRoomDefensePanel(panelRoomDefense)
		UpdateRoomItemsPanel(panelRoomItems)
	}
	
	// Initial panel update (must be before AddLog)
	updatePanels()
	
	// Initial log messages
	AddLog(panelLog, "[green]Welcome to Haunted Room Defense![white]")
	AddLog(panelLog, "[cyan]Defend your room from Dream Hunters![white]")
	AddLog(panelLog, "[yellow]Buy beds to generate coins![white]")
	updatePanels()

	// Bottom row: Room Defense and Room Items side by side
	bottomRow := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(panelRoomDefense, 0, 1, false).
		AddItem(panelRoomItems, 0, 1, false)
	bottomRow.SetBorderPadding(0, 0, 0, 0)

	// Left column: Logs on top, bottom row below
	leftColumn := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(panelLog, 0, 2, false).
		AddItem(bottomRow, 0, 1, false)
	leftColumn.SetBorderPadding(0, 0, 0, 0)

	// Right column: Your Items and Shop stacked vertically
	rightColumn := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(panelYourItems, 0, 1, false).
		AddItem(panelShop, 0, 2, false)
	rightColumn.SetBorderPadding(0, 0, 0, 0)

	// Create main layout: left column and right column side by side
	flex := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(leftColumn, 0, 2, false).
		AddItem(rightColumn, 0, 1, true)
	flex.SetBorderPadding(0, 0, 0, 0)

	// Game loop ticker
	ticker := time.NewTicker(1 * time.Second)
	combatTicker := time.NewTicker(100 * time.Millisecond) // Combat updates 10x per second

	// Create game over modal (without done func yet)
	gameOverModal := tview.NewModal().
		SetText("").
		AddButtons([]string{"Yes", "No"}).
		SetBackgroundColor(tcell.ColorBlack).
		SetButtonBackgroundColor(tcell.ColorBlack).
		SetButtonTextColor(tcell.ColorWhite).
		SetTextColor(tcell.ColorWhite)

	// Pages to handle modal overlay
	pages := tview.NewPages().
		AddPage("main", flex, true, true).
		AddPage("gameOver", gameOverModal, true, false)
	
	// Set modal done function (now that pages is declared)
	gameOverModal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		if buttonLabel == "Yes" {
			// Restart game
			InitGame()
			selectedItem = 0
			shopCategory = 0
			pages.HidePage("gameOver")
			updatePanels()
		} else {
			// Quit
			ticker.Stop()
			combatTicker.Stop()
			app.Stop()
		}
	})
	
	go func() {
		hunterSpawnCounter := 0
		for range ticker.C {
			UpdateGame()
			
			// Check for game over
			gs := GetGameState()
			if gs.gameOver {
				if gs.gameWon {
					gameOverModal.SetText("🎉 VICTORY! 🎉\nYou defeated the Dream Hunter!\n\nPlay again?")
				} else {
					gameOverModal.SetText("💀 GAME OVER 💀\nYour door was destroyed!\n\nPlay again?")
				}
				pages.ShowPage("gameOver")
			}
			
			// Spawn hunter every 10 seconds
			hunterSpawnCounter++
			if hunterSpawnCounter >= 10 {
				SpawnHunter(panelLog)
				hunterSpawnCounter = 0
			}
			
			updatePanels()
		}
	}()
	
	go func() {
		for range combatTicker.C {
			UpdateCombat(panelLog)
			updatePanels()
		}
	}()

	// Global keyboard shortcuts
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// If game over modal is visible, let it handle input
		gs := GetGameState()
		if gs.gameOver {
			return event
		}
		
		items := GetAvailableItemsByCategory(shopCategory)
		
		switch event.Key() {
		case tcell.KeyCtrlC:
			// Exit application
			ticker.Stop()
			combatTicker.Stop()
			app.Stop()
			return nil
		case tcell.KeyUp:
			// Move selection up
			if selectedItem > 0 {
				selectedItem--
				updatePanels()
			}
			return nil
		case tcell.KeyDown:
			// Move selection down
			if selectedItem < len(items)-1 {
				selectedItem++
				updatePanels()
			}
			return nil
		case tcell.KeyLeft:
			// Previous category
			if shopCategory > 0 {
				shopCategory--
				selectedItem = 0
				updatePanels()
			}
			return nil
		case tcell.KeyRight:
			// Next category
			if shopCategory < 2 {
				shopCategory++
				selectedItem = 0
				updatePanels()
			}
			return nil
		}
		
		// Handle character keys
		switch event.Rune() {
		case 'i', 'I':
			// Buy selected item
			BuyItemByCategory(selectedItem, shopCategory, panelLog)
			updatePanels()
			return nil
		case 'h', 'H':
			// Spawn hunter manually for testing
			SpawnHunter(panelLog)
			updatePanels()
			return nil
		case 'q', 'Q':
			// Quit
			ticker.Stop()
			combatTicker.Stop()
			app.Stop()
			return nil
		}
		
		return event
	})

	// Run the application
	if err := app.SetRoot(pages, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
