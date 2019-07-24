package ipa

type InfoPlist struct {
	BundleShortVersion 	string 	`plist:"CFBundleShortVersionString"`
	BundleVersion 	 	string 	`plist:"CFBundleVersion"`
	BundleName 			string 	`plist:"CFBundleName"`
	BundleId 			string	`plist:"CFBundleIdentifier"`
}