package esxi

import (
	"fmt"

	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"golang.org/x/net/context"
)

// 内容库操作
func GetLibraryItem(ctx context.Context, rc *rest.Client, libraryName string,
	libraryItemType string, libraryItemName string) (*library.Item, error) {
	// 需要通过 rc 来获得 library.Manager 对象
	m := library.NewManager(rc)
	// 通过内容库的名称来查找内容库
	libraries, err := m.FindLibrary(ctx, library.Find{Name: libraryName})
	if err != nil {
		fmt.Printf("Find library by name %s failed, %v", libraryName, err)
		return nil, err
	}

	// 判断是否找到
	if len(libraries) == 0 {
		fmt.Printf("Library %s was not found", libraryName)
		return nil, fmt.Errorf("library %s was not found", libraryName)
	}

	if len(libraries) > 1 {
		fmt.Printf("There are multiple libraries with the name %s", libraryName)
		return nil, fmt.Errorf("there are multiple libraries with the name %s", libraryName)
	}

	// 在内容库中通过 ovf 模板的名称来找到 ovf 模板
	items, err := m.FindLibraryItems(ctx, library.FindItem{Name: libraryItemName,
		Type: libraryItemType, LibraryID: libraries[0]})

	if err != nil {
		fmt.Printf("Find library item by name %s failed", libraryItemName)
		return nil, fmt.Errorf("find library item by name %s failed", libraryItemName)
	}

	if len(items) == 0 {
		fmt.Printf("Library item %s was not found", libraryItemName)
		return nil, fmt.Errorf("library item %s was not found", libraryItemName)
	}

	if len(items) > 1 {
		fmt.Printf("There are multiple library items with the name %s", libraryItemName)
		return nil, fmt.Errorf("there are multiple library items with the name %s", libraryItemName)
	}

	item, err := m.GetLibraryItem(ctx, items[0])
	if err != nil {
		fmt.Printf("Get library item by %s failed, %v", items[0], err)
		return nil, err
	}

	return item, nil
}
