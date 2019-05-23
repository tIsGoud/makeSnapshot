// Copyright © 2019 Albert W. Alberts <a.w.alberts@tisgoud.nl>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

// GetBearerTokenRequest ...
type GetBearerTokenRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Tenant   string `json:"tenant"`
}

// GetBearerTokenResponse ...
type GetBearerTokenResponse struct {
	Expires string `json:"expires"`
	ID      string `json:"id"`
	Tenant  string `json:"tenant"`
}

// SnapShotTemplate ...
type SnapShotTemplate struct {
	Type        string      `json:"type"`
	ResourceID  string      `json:"resourceId"`
	ActionID    string      `json:"actionId"`
	Description interface{} `json:"description"`
	Data        Data        `json:"data"`
}

// Data ...
type Data struct {
	ProviderASDPRESENTATIONINSTANCE interface{} `json:"provider-__ASD_PRESENTATION_INSTANCE"`
	ProviderAsdTenantRef            string      `json:"provider-__asd_tenantRef"`
	ProviderDeleteExisting          interface{} `json:"provider-deleteExisting"`
	ProviderDescription             interface{} `json:"provider-description"`
	ProviderExistingSnapshotName    interface{} `json:"provider-existingSnapshotName"`
	ProviderName                    interface{} `json:"provider-name"`
}
