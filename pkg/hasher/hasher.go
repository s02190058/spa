package hasher

import "golang.org/x/crypto/bcrypt"

type Hasher struct {
	cost int
}

func New(cost int) *Hasher {
	return &Hasher{
		cost: cost,
	}
}

func (h *Hasher) Encrypt(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), h.cost)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (h *Hasher) Compare(encryptedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password)) == nil
}
