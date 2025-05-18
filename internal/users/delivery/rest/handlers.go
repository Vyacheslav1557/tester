package rest

import (
	"context"
	testerv1 "github.com/Vyacheslav1557/tester/contracts/tester/v1"
	"github.com/Vyacheslav1557/tester/internal/models"
	"github.com/Vyacheslav1557/tester/internal/users"
	"github.com/Vyacheslav1557/tester/pkg"
	"github.com/gofiber/fiber/v2"
)

type Handlers struct {
	usersUC users.UseCase
}

func NewHandlers(usersUC users.UseCase) *Handlers {
	return &Handlers{
		usersUC: usersUC,
	}
}

const (
	sessionKey = "session"
)

func sessionFromCtx(ctx context.Context) (*models.Session, error) {
	const op = "sessionFromCtx"

	session, ok := ctx.Value(sessionKey).(*models.Session)
	if !ok {
		return nil, pkg.Wrap(pkg.ErrUnauthenticated, nil, op, "")
	}

	return session, nil
}

func (h *Handlers) CreateUser(c *fiber.Ctx) error {
	const op = "UsersHandlers.CreateUser"

	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		var req = &testerv1.CreateUserRequest{}
		err := c.BodyParser(req)
		if err != nil {
			return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to parse request")
		}

		id, err := h.usersUC.CreateUser(ctx,
			&models.UserCreation{
				Username: req.Username,
				Password: req.Password,
				Role:     models.RoleStudent,
			},
		)
		if err != nil {
			return err
		}

		return c.JSON(testerv1.CreationResponse{Id: id})
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) GetUser(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher, models.RoleStudent:
		user, err := h.usersUC.ReadUserById(c.Context(), id)
		if err != nil {
			return err
		}

		return c.JSON(testerv1.GetUserResponse{
			User: UserDTO(*user),
		})
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) UpdateUser(c *fiber.Ctx, id int32) error {
	const op = "UsersHandlers.UpdateUser"
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin:
		var req = &testerv1.UpdateUserRequest{}
		err := c.BodyParser(req)
		if err != nil {
			return pkg.Wrap(pkg.ErrBadInput, err, op, "failed to parse request")
		}

		err = h.usersUC.UpdateUser(ctx, id, &models.UserUpdate{
			Username: req.Username,
			Role:     RoleDTO(req.Role),
		})
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) DeleteUser(c *fiber.Ctx, id int32) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	switch session.Role {
	case models.RoleAdmin:
		ctx := c.Context()

		err := h.usersUC.DeleteUser(ctx, id)
		if err != nil {
			return err
		}

		return c.SendStatus(fiber.StatusOK)
	default:
		return pkg.NoPermission
	}
}

func (h *Handlers) ListUsers(c *fiber.Ctx, params testerv1.ListUsersParams) error {
	ctx := c.Context()

	session, err := sessionFromCtx(ctx)
	if err != nil {
		return err
	}

	filters := models.UsersListFilters{
		PageSize: params.PageSize,
		Page:     params.Page,
		Role:     RoleDTO(params.Role),
		Username: params.Username,
		Order:    params.Order,
	}

	switch session.Role {
	case models.RoleAdmin, models.RoleTeacher:
		usersList, err := h.usersUC.ListUsers(c.Context(), filters)
		if err != nil {
			return err
		}

		resp := testerv1.ListUsersResponse{
			Users:      make([]testerv1.User, len(usersList.Users)),
			Pagination: PaginationDTO(usersList.Pagination),
		}

		for i, user := range usersList.Users {
			resp.Users[i] = UserDTO(*user)
		}

		return c.JSON(resp)
	default:
		return pkg.NoPermission
	}
}

func RoleDTO(i *int32) *models.Role {
	if i == nil {
		return nil
	}
	ii := models.Role(*i)
	return &ii
}

func PaginationDTO(p models.Pagination) testerv1.Pagination {
	return testerv1.Pagination{
		Page:  p.Page,
		Total: p.Total,
	}
}

// UserDTO sanitizes password
func UserDTO(u models.User) testerv1.User {
	return testerv1.User{
		Id:        u.Id,
		Username:  u.Username,
		Role:      int32(u.Role),
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
