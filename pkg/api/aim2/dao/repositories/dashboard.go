package repositories

// Service provides service layer to work with `dashboard` business logic.
type Service struct{}

// NewService creates new Service instance.
func NewService() *Service {
	return &Service{}
}

func (s Service) GetDashboards() {
	var dashboards []database.Dashboard
	if err := database.DB.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		Order(clause.OrderByColumn{
			Column: clause.Column{
				Table: "App",
				Name:  "updated_at",
			},
			Desc: true,
		}).
		Find(&dashboards).
		Error; err != nil {
		return fmt.Errorf("error fetching dashboards: %w", err)
	}
}

func (s Service) Create() {
	app := database.App{
		Base: database.Base{
			ID: req.AppID,
		},
		NamespaceID: ns.ID,
	}
	if err := database.DB.
		Select("ID", "Type").
		Where("NOT is_archived").
		First(&app).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", req.AppID, err))
	}

	dash := database.Dashboard{
		AppID:       &req.AppID,
		App:         app,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := database.DB.
		Omit("App").
		Create(&dash).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error inserting dashboard: %s", err))
	}
}

func (s Service) Get() {
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dashboard := database.Dashboard{
		Base: database.Base{
			ID: req.ID,
		},
	}
	if err := database.DB.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		First(&dashboard).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find dashboard %q: %s", req.ID, err))
	}
}

func (s Service) Update() {
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	if err := ctx.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dash := database.Dashboard{
		Base: database.Base{
			ID: req.ID,
		},
	}
	if err := database.DB.
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		First(&dash).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find dashboard %q: %s", req.ID, err))
	}

	if err := database.DB.
		Omit("App").
		Model(&dash).
		Updates(database.Dashboard{
			Name:        req.Name,
			Description: req.Description,
		}).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("error updating dashboard %q: %s", req.ID, err))
	}
}

func (s Service) Delete() {
	if err := ctx.ParamsParser(&req); err != nil {
		return fiber.NewError(fiber.StatusUnprocessableEntity, err.Error())
	}

	dash := database.Dashboard{
		Base: database.Base{
			ID: req.ID,
		},
	}
	if err := database.DB.
		Select("dashboards.id").
		InnerJoins(
			"App",
			database.DB.Select(
				"ID", "Type",
			).Where(
				&database.App{
					NamespaceID: ns.ID,
				},
				"NamespaceID",
			),
		).
		Where("NOT dashboards.is_archived").
		First(&dash).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.ErrNotFound
		}
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to find app %q: %s", req.ID, err))
	}

	if err := database.DB.
		Omit("App").
		Model(&dash).
		Update("IsArchived", true).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("unable to delete app %q: %s", req.ID, err))
	}
}


