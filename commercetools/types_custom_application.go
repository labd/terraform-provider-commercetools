package commercetools

// LabelLocale describes the structure of a localized label.
type LabelLocale struct {
	Locale string `json:"locale"`
	Value  string `json:"value"`
}

// NavbarSubmenu describes the structure of a Custom Application navigation submenu stored object.
type NavbarSubmenu struct {
	Key             string        `json:"key"`
	URIPath         string        `json:"uriPath"`
	LabelAllLocales []LabelLocale `json:"labelAllLocales"`
	Permissions     []string      `json:"permissions"`
}

// NavbarMenu describes the structure of a Custom Application navigation menu stored object.
type NavbarMenu struct {
	Key             string          `json:"key"`
	URIPath         string          `json:"uriPath"`
	Icon            string          `json:"icon"`
	LabelAllLocales []LabelLocale   `json:"labelAllLocales"`
	Permissions     []string        `json:"permissions"`
	Submenu         []NavbarSubmenu `json:"submenu"`
}

// CustomApplication describes the structure of a Custom Application stored object.
type CustomApplication struct {
	ID          string     `json:"id"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
	IsActive    bool       `json:"isActive"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	URL         string     `json:"url"`
	NavbarMenu  NavbarMenu `json:"navbarMenu"`
}

// ProjectExtension describes the structure of a project extension stored object.
type ProjectExtension struct {
	ID           string              `json:"id"`
	Applications []CustomApplication `json:"applications"`
}

// GraphQLResponseProjectExtension describes the structure of the query result for fetching a Custom Application.
type GraphQLResponseProjectExtension struct {
	ProjectExtension *ProjectExtension `json:"projectExtension"`
}

// GraphQLResponseProjectExtensionCreation describes the structure of the query result for creating a Custom Application.
type GraphQLResponseProjectExtensionCreation struct {
	CreateProjectExtensionApplication *ProjectExtension `json:"createProjectExtensionApplication"`
}

// GraphQLResponseProjectExtensionUpdate describes the structure of the query result for updating a Custom Application.
type GraphQLResponseProjectExtensionUpdate struct {
	UpdateProjectExtensionApplication     *ProjectExtension `json:"updateProjectExtensionApplication"`
	ActivateProjectExtensionApplication   *ProjectExtension `json:"activateProjectExtensionApplication"`
	DeactivateProjectExtensionApplication *ProjectExtension `json:"deactivateProjectExtensionApplication"`
}

// GraphQLResponseProjectExtensionDeletion describes the structure of the query result for deleting a Custom Application.
type GraphQLResponseProjectExtensionDeletion struct {
	DeleteProjectExtensionApplication *ProjectExtension `json:"deleteProjectExtensionApplication"`
}
