package filters

import (
  "strings"
)

func ParseFilters(filtersDefinition string) (map[string][]string, error) {
	parsedFilters := make(map[string][]string)
	for _, section := range strings.Split(filtersDefinition, "|") {
		if section != "" {
      filter := strings.Split(section, ":")
			filterType := strings.TrimSpace(filter[0])
			filterDef  := strings.TrimSpace(filter[1])
      filterValues := []string{}
      for _, filterValue := range strings.Split(filterDef, ",") {
        filterValues = append(filterValues, strings.TrimSpace(filterValue))
      }
      parsedFilters["cf_" + filterType] = filterValues
	  }
  }
	return parsedFilters, nil
}
