package services

import (
	"errors"

	"config-center/internal/database"
	"config-center/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Login(username, password string) (*models.User, error) {
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	return &user, nil
}

func (s *AuthService) CreateUser(username, password string) (*models.User, error) {
	var existing models.User
	if err := database.DB.Where("username = ?", username).First(&existing).Error; err == nil {
		return nil, errors.New("username already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hash),
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ListUsers() ([]models.User, error) {
	var users []models.User
	err := database.DB.Find(&users).Error
	return users, err
}

func (s *AuthService) DeleteUser(id uint) error {
	return database.DB.Delete(&models.User{}, id).Error
}

type RoleService struct{}

func NewRoleService() *RoleService {
	return &RoleService{}
}

var roleHierarchy = map[string]int{
	models.RoleViewer: 1,
	models.RoleEditor: 2,
	models.RoleAdmin:  3,
}

func (s *RoleService) GetUserRoles(userID uint) ([]models.UserRole, error) {
	var roles []models.UserRole
	err := database.DB.Where("user_id = ?", userID).Find(&roles).Error
	return roles, err
}

func (s *RoleService) GetHighestRoleForNamespace(userID uint, namespaceID uint) string {
	var roles []models.UserRole
	database.DB.Where("user_id = ?", userID).Find(&roles)

	highestLevel := 0
	highestRole := ""

	for _, r := range roles {
		if r.NamespaceID == nil || *r.NamespaceID == namespaceID {
			level := roleHierarchy[r.Role]
			if level > highestLevel {
				highestLevel = level
				highestRole = r.Role
			}
		}
	}

	return highestRole
}

func (s *RoleService) HasPermission(userID uint, namespaceID uint, requiredRole string) bool {
	userRole := s.GetHighestRoleForNamespace(userID, namespaceID)
	userLevel := roleHierarchy[userRole]
	requiredLevel := roleHierarchy[requiredRole]
	return userLevel >= requiredLevel
}

func (s *RoleService) GetAccessibleNamespaceIDs(userID uint) []uint {
	var roles []models.UserRole
	database.DB.Where("user_id = ?", userID).Find(&roles)

	for _, r := range roles {
		if r.Role == models.RoleAdmin && r.NamespaceID == nil {
			return nil
		}
	}

	var nsIDs []uint
	seen := make(map[uint]bool)
	for _, r := range roles {
		if r.NamespaceID != nil && !seen[*r.NamespaceID] {
			nsIDs = append(nsIDs, *r.NamespaceID)
			seen[*r.NamespaceID] = true
		}
	}
	return nsIDs
}

func (s *RoleService) IsGlobalAdmin(userID uint) bool {
	var count int64
	database.DB.Model(&models.UserRole{}).
		Where("user_id = ? AND role = ? AND namespace_id IS NULL", userID, models.RoleAdmin).
		Count(&count)
	return count > 0
}

func (s *RoleService) GrantRole(userID uint, namespaceID *uint, role string) error {
	ur := &models.UserRole{
		UserID:      userID,
		NamespaceID: namespaceID,
		Role:        role,
	}
	return database.DB.Create(ur).Error
}

func (s *RoleService) RevokeRole(id uint) error {
	return database.DB.Delete(&models.UserRole{}, id).Error
}

func (s *RoleService) GetRoleBindings(userID uint) ([]models.UserRole, error) {
	var roles []models.UserRole
	err := database.DB.Where("user_id = ?", userID).Find(&roles).Error
	return roles, err
}
