package discovery

import (
	"strings"

	"github.com/codeshelldev/gotl/pkg/logger"
)

type Diff[T any] struct {
	Added	[]T
	Removed	[]T
}

func (diff *Diff[T]) Merge(other Diff[T]) {
	diff.Added = append(diff.Added, other.Added...)
	diff.Removed = append(diff.Removed, other.Removed...)
}

func CleanDiff[T comparable](diff Diff[T]) Diff[T] {
    removedMap := map[T]struct{}{}
    addedMap := map[T]struct{}{}

    for _, r := range diff.Removed {
        removedMap[r] = struct{}{}
    }
    for _, a := range diff.Added {
        addedMap[a] = struct{}{}
    }

    // Remove items that are in both
    for removed := range removedMap {
		_, exists := addedMap[removed]
        if exists {
            delete(removedMap, removed)
            delete(addedMap, removed)
        }
    }

    cleaned := Diff[T]{}

    for removed := range removedMap {
        cleaned.Removed = append(cleaned.Removed, removed)
    }
    for added := range addedMap {
        cleaned.Added = append(cleaned.Added, added)
    }

	return cleaned
}

func logDiff(id string, diff Diff[string]) {
	if len(diff.Added) <= 0 && len(diff.Removed) <= 0 {
		return
	}

	addedStr := strings.Join(diff.Added, ",")
	removedStr := strings.Join(diff.Removed, ",")

	logger.Debug("[", id, "] ", "(+) ", addedStr, " (-) ", removedStr)
}

func GetDiff[T comparable](old, new []T) Diff[T] {
	diff := Diff[T]{}

	oldMap := map[T]struct{}{}
	newMap := map[T]struct{}{}

	for _, value := range old {
		oldMap[value] = struct{}{}
	}
	for _, value := range new {
		newMap[value] = struct{}{}
	}

	for value := range oldMap {
		_, exists := newMap[value]

		if !exists {
			diff.Removed = append(diff.Removed, value)
		}
	}
	for value := range newMap {
		_, exists := oldMap[value]; 

		if !exists {
			diff.Added = append(diff.Added, value)
		}
	}

	return diff
}