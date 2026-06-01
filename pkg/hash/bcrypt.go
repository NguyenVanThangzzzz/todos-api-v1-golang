package hash

import "golang.org/x/crypto/bcrypt"

// Password băm mật khẩu thuần bằng bcrypt (đã kèm salt tự sinh).
// Lưu ý: bcrypt chỉ xử lý tối đa 72 byte đầu của mật khẩu.
func Password(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Compare so khớp mật khẩu thuần với hash đã lưu. Trả về true nếu khớp.
func Compare(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
