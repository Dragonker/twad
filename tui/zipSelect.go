package tui

import (
	"os"
	"path"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/Dragonker/twad/base"
	"github.com/Dragonker/twad/helper"
)

type zipImportUI struct {
	layout                *tview.Flex
	selectTree            *tview.TreeView
	modNameInput          *tview.InputField
	modNameForm           *tview.Form
	importSecurityWarning *tview.TextView

	zipPath string
	modName string
}

func newZipImportUI() *zipImportUI {
	var zui zipImportUI
	zui.initZipSelect()
	zui.initZipImportForm("")
	zui.importSecurityWarning = tview.NewTextView().SetText(dict.zipImportSecurityWarn).SetTextColor(tcell.ColorRed)

	zui.layout = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(zui.importSecurityWarning, 1, 0, true).
		AddItem(zui.selectTree, 0, 1, true).
		AddItem(zui.modNameForm, 7, 0, false)
	return &zui
}

func (z *zipImportUI) initZipSelect() {
	//rootDir := helper.Home() // TODO: Start from / but preselect /home/user
	rootDir := "/"
	if _, err := os.Stat(rootDir); err != nil {
		if os.IsNotExist(err) {
			// TODO
		}
	}

	var rootNode *tview.TreeNode
	z.selectTree, rootNode = newTree(rootDir)
	z.selectTree.SetTitle(dict.zipSelectTitle)
	add := makeFileTreeAddFunc(helper.FilterExtensions, ".zip.tar.gz.rar", true, true)
	add(rootNode, rootDir)

	z.selectTree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()

		if reference == nil {
			return
		}
		children := node.GetChildren()
		if len(children) == 0 {
			// Load and show files in this directory.
			selPath := reference.(string)

			// check if path can at leas be read
			// otherwise return
			f, err := os.OpenFile(selPath, os.O_RDONLY, 0666)
			if err != nil && os.IsPermission(err) {
				return
			}
			defer f.Close()

			fi, err := os.Stat(selPath)
			switch {
			case err != nil:
				return // TODO: any form of info to user?
			case fi.IsDir():
				add(node, selPath)
			default:
				z.zipPath = selPath
				z.modNameInput.SetText(strings.TrimSuffix(path.Base(selPath), path.Ext(selPath)))
				app.SetFocus(z.modNameForm)
			}
		} else {
			node.SetExpanded(!node.IsExpanded())
		}
	})

	z.selectTree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		if k == tcell.KeyRune {
			switch event.Rune() {
			case 'q':
				app.Stop()
				return nil
			}
		}
		return event
	})
}

func (z *zipImportUI) initZipImportForm(archivePath string) {
	z.modNameInput = tview.NewInputField().SetLabel(dict.zipImportToLabel).SetText(path.Base(archivePath))
	if archivePath == "" {
		z.modNameInput.SetText("")
	}

	modNameDoneCheck := func() {
		suggestedName := z.modNameInput.GetText()
		if !helper.IsFileNameValid(suggestedName) {
			z.modNameInput.SetLabel(dict.zipImportToLabel + warnColor + " " + dict.zipImportToBadNameLabel)
			return
		}
		if _, err := os.Stat(path.Join(base.Config().WadDir, suggestedName)); !os.IsNotExist(err) {
			z.modNameInput.SetLabel(dict.zipImportToLabel + warnColor + " " + dict.zipImportToExistsLabel)
			return
		}
		z.modNameInput.SetLabel(dict.zipImportToLabel)
	}

	z.modNameInput.SetDoneFunc(func(key tcell.Key) {
		modNameDoneCheck()
	})

	z.modNameForm = tview.NewForm().
		AddFormItem(z.modNameInput).
		AddButton(dict.zipImportFormOk, func() {
			z.modName = z.modNameInput.GetText()

			// test file name again
			if !helper.IsFileNameValid(z.modName) {
				showError("Cannot use that name", "Possible reasons:\n- File name contains forbidden characters\n- No permission to write this file/folder", z.modNameInput, nil)
				return
			}

			// test if provided zip exists
			if _, err := os.Stat(z.zipPath); os.IsNotExist(err) {
				showError("Mod archive not found", err.Error(), zipInput.selectTree, nil)
				zipInput.reset()
				return
			}

			// START ACTUAL IMPORT
			if err := base.ImportArchive(z.zipPath, z.modName); err != nil {
				showError("Could not import zip", err.Error(), zipInput.selectTree, nil)
			}
			z.reset()
		}).
		AddButton(dict.zipImportCancel, func() {
			z.reset()
		})

	z.modNameForm.
		SetBorder(true).
		SetTitle(dict.zipImportFormTitle)
	z.modNameForm.SetFocus(0)
}

func (z *zipImportUI) reset() {
	z.modNameInput.SetText("").SetLabel(dict.zipImportToLabel)
	z.modNameForm.SetFocus(0)
	z.modName = ""
	z.zipPath = ""
	app.SetFocus(z.selectTree)
}
