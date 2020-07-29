package authcode

import (
	"context"

	"github.com/int128/kubelogin/pkg/adaptors/logger"
	"github.com/int128/kubelogin/pkg/adaptors/oidcclient"
	"github.com/int128/kubelogin/pkg/adaptors/reader"
	"github.com/int128/kubelogin/pkg/domain/oidc"
	"github.com/int128/kubelogin/pkg/domain/pkce"
	"golang.org/x/xerrors"
)

const keyboardPrompt = "Enter code: "
const oobRedirectURI = "urn:ietf:wg:oauth:2.0:oob"

type KeyboardOption struct {
	AuthRequestExtraParams map[string]string
}

// Keyboard provides the authorization code flow with keyboard interactive.
type Keyboard struct {
	Reader reader.Interface
	Logger logger.Interface
}

func (u *Keyboard) Do(ctx context.Context, o *KeyboardOption, client oidcclient.Interface) (*oidc.TokenSet, error) {
	u.Logger.V(1).Infof("starting the authorization code flow with keyboard interactive")
	state, err := oidc.NewState()
	if err != nil {
		return nil, xerrors.Errorf("could not generate a state: %w", err)
	}
	nonce, err := oidc.NewNonce()
	if err != nil {
		return nil, xerrors.Errorf("could not generate a nonce: %w", err)
	}
	p, err := pkce.New(client.SupportedPKCEMethods())
	if err != nil {
		return nil, xerrors.Errorf("could not generate PKCE parameters: %w", err)
	}
	authCodeURL := client.GetAuthCodeURL(oidcclient.AuthCodeURLInput{
		State:                  state,
		Nonce:                  nonce,
		PKCEParams:             p,
		RedirectURI:            oobRedirectURI,
		AuthRequestExtraParams: o.AuthRequestExtraParams,
	})
	u.Logger.Printf("Please visit the following URL in your browser: %s", authCodeURL)
	code, err := u.Reader.ReadString(keyboardPrompt)
	if err != nil {
		return nil, xerrors.Errorf("could not read an authorization code: %w", err)
	}

	u.Logger.V(1).Infof("exchanging the code and token")
	tokenSet, err := client.ExchangeAuthCode(ctx, oidcclient.ExchangeAuthCodeInput{
		Code:        code,
		PKCEParams:  p,
		Nonce:       nonce,
		RedirectURI: oobRedirectURI,
	})
	if err != nil {
		return nil, xerrors.Errorf("could not exchange the authorization code: %w", err)
	}
	u.Logger.V(1).Infof("finished the authorization code flow with keyboard interactive")
	return &oidc.TokenSet{
		IDToken:       tokenSet.IDToken,
		IDTokenClaims: tokenSet.IDTokenClaims,
		RefreshToken:  tokenSet.RefreshToken,
	}, nil
}