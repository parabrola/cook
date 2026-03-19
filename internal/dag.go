package internal

import "fmt"

func ValidateDependencies(tasks taskList) error {
	for name, task := range tasks {
		for _, dep := range task.DependsOn {
			if _, ok := tasks[dep]; !ok {
				return fmt.Errorf("task '%s' depends on unknown task '%s'", name, dep)
			}
		}
	}

	return detectCycle(tasks)
}

func detectCycle(tasks taskList) error {
	const (
		unvisited = 0
		visiting  = 1
		visited   = 2
	)

	state := make(map[string]int, len(tasks))
	var path []string

	var visit func(name string) error
	visit = func(name string) error {
		if state[name] == visited {
			return nil
		}
		if state[name] == visiting {
			cycle := append(path, name)
			start := 0
			for i, n := range cycle {
				if n == name {
					start = i
					break
				}
			}
			return fmt.Errorf("circular dependency detected: %s", formatCycle(cycle[start:]))
		}

		state[name] = visiting
		path = append(path, name)

		task := tasks[name]
		for _, dep := range task.DependsOn {
			if err := visit(dep); err != nil {
				return err
			}
		}

		path = path[:len(path)-1]
		state[name] = visited
		return nil
	}

	for name := range tasks {
		if state[name] == unvisited {
			if err := visit(name); err != nil {
				return err
			}
		}
	}

	return nil
}

func formatCycle(cycle []string) string {
	result := cycle[0]
	for i := 1; i < len(cycle); i++ {
		result += " -> " + cycle[i]
	}
	return result
}
