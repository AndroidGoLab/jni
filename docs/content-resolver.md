# ContentResolver and Content Providers

The `content/resolver` package wraps `android.content.ContentResolver` for querying content providers (contacts, settings, media, etc.) from Go.

## Getting a ContentResolver

```go
import "github.com/AndroidGoLab/jni/content/resolver"

ctx, _ := app.ContextFromObject(vm, activityRef)
defer ctx.Close()

resolverObj, err := ctx.GetContentResolver()
if err != nil {
    return fmt.Errorf("GetContentResolver: %w", err)
}
cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}
```

## Parsing Content URIs

```go
uriHelper := resolver.Uri{VM: vm}
uri, err := uriHelper.Parse("content://settings/system")
```

Common content URIs:
- `content://settings/system` — System settings (no permission needed)
- `content://settings/secure` — Secure settings
- `content://com.android.contacts/contacts` — Contacts (requires `READ_CONTACTS`)
- `content://media/external/images/media` — Media images (requires `READ_MEDIA_IMAGES`)

## Querying with a Cursor

```go
// Query returns a Cursor for iterating rows
cursorObj, err := cr.Query4(uri, nil, nil, nil)
if err != nil {
    return fmt.Errorf("query: %w", err)
}
if cursorObj == nil {
    fmt.Println("No results")
    return nil
}

cursor := resolver.Cursor{VM: vm, Obj: cursorObj}
defer cursor.Close()

// Get row and column counts
count, _ := cursor.GetCount()
colCount, _ := cursor.GetColumnCount()
fmt.Printf("Rows: %d, Columns: %d\n", count, colCount)

// Print column names
for c := int32(0); c < colCount; c++ {
    name, _ := cursor.GetColumnName(c)
    fmt.Printf("  col[%d]: %s\n", c, name)
}
```

## Iterating Rows

```go
// Move to the first row
ok, err := cursor.MoveToFirst()
if err != nil || !ok {
    return nil
}

for {
    // Read string values by column index
    for c := int32(0); c < colCount; c++ {
        val, err := cursor.GetString(c)
        if err != nil {
            val = "(error)"
        }
        fmt.Printf("%s ", val)
    }
    fmt.Println()

    // Move to next row
    moved, err := cursor.MoveToNext()
    if err != nil || !moved {
        break
    }
}
```

## Complete Example: Query System Settings

```go
import (
    "github.com/AndroidGoLab/jni"
    "github.com/AndroidGoLab/jni/app"
    "github.com/AndroidGoLab/jni/content/resolver"
)

func querySettings(vm *jni.VM, activityRef *jni.GlobalRef) error {
    ctx, _ := app.ContextFromObject(vm, activityRef)
    defer ctx.Close()

    resolverObj, _ := ctx.GetContentResolver()
    cr := resolver.ContentResolver{VM: vm, Obj: resolverObj}

    uriHelper := resolver.Uri{VM: vm}
    uri, _ := uriHelper.Parse("content://settings/system")

    cursorObj, err := cr.Query4(uri, nil, nil, nil)
    if err != nil {
        return err
    }
    if cursorObj == nil {
        return nil
    }

    cursor := resolver.Cursor{VM: vm, Obj: cursorObj}
    defer cursor.Close()

    count, _ := cursor.GetCount()
    colCount, _ := cursor.GetColumnCount()
    fmt.Printf("Settings: %d rows, %d columns\n", count, colCount)

    ok, _ := cursor.MoveToFirst()
    for ok {
        var vals []string
        for c := int32(0); c < colCount; c++ {
            s, _ := cursor.GetString(c)
            vals = append(vals, s)
        }
        fmt.Println(vals)

        ok, _ = cursor.MoveToNext()
    }
    return nil
}
```

## Cursor Methods

The `Cursor` type wraps `android.database.Cursor` with these methods:

| Method | Description |
|--------|-------------|
| `GetCount()` | Total number of rows |
| `GetColumnCount()` | Number of columns |
| `GetColumnName(index)` | Column name by index |
| `GetColumnIndex(name)` | Column index by name |
| `MoveToFirst()` | Move cursor to first row |
| `MoveToNext()` | Move cursor to next row |
| `MoveToPosition(pos)` | Move cursor to specific row |
| `GetString(colIndex)` | Read string value from column |
| `GetInt(colIndex)` | Read int value from column |
| `GetLong(colIndex)` | Read long value from column |
| `IsNull(colIndex)` | Check if column value is null |
| `Close()` | Release the cursor |

## Querying Contacts

```go
// Requires android.permission.READ_CONTACTS
uri, _ := uriHelper.Parse("content://com.android.contacts/contacts")
cursorObj, err := cr.Query4(uri, nil, nil, nil)
// ... iterate as above, reading display_name, phone number, etc.
```

## Required Permissions

Permissions depend on which content provider you query:
- System settings: no permission needed
- Contacts: `READ_CONTACTS`
- Media: `READ_MEDIA_IMAGES`, `READ_MEDIA_VIDEO`, `READ_MEDIA_AUDIO` (Android 13+)
- Call log: `READ_CALL_LOG`
