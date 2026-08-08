package account

// AccountType identifies the channel or bank that owns an account.
type AccountType string

const (
	AccountTypeWeChat AccountType = "WeChat"
	AccountTypeAliPay AccountType = "AliPay"
	AccountTypeBOC    AccountType = "BOC"
)

func (t AccountType) IsValid() bool {
	switch t {
	case AccountTypeWeChat, AccountTypeAliPay, AccountTypeBOC:
		return true
	default:
		return false
	}
}
