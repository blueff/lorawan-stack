// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

func init() {
	const forwards = `
		CREATE TABLE IF NOT EXISTS applications (
			application_id   STRING(36) PRIMARY KEY,
			description      TEXT,
			created_at       TIMESTAMP DEFAULT current_timestamp(),
			updated_at       TIMESTAMP DEFAULT current_timestamp()
		);

		CREATE TABLE IF NOT EXISTS applications_api_keys (
			application_id   STRING(36) NOT NULL REFERENCES applications(application_id),
			key              STRING PRIMARY KEY,
			key_name         STRING(36) NOT NULL,
			UNIQUE(application_id, key_name)
		);

		CREATE TABLE IF NOT EXISTS applications_api_keys_rights (
			key       STRING NOT NULL REFERENCES applications_api_keys(key),
			"right"   STRING NOT NULL,
			PRIMARY KEY(key, "right")
		);

		CREATE TABLE IF NOT EXISTS applications_collaborators (
			application_id   STRING(36) REFERENCES applications(application_id),
			user_id          STRING(36) REFERENCES users(user_id),
			"right"          STRING NOT NULL,
			PRIMARY KEY(application_id, user_id, "right")
		);
	`

	const backwards = `
		DROP TABLE IF EXISTS applications_collaborators;
		DROP TABLE IF EXISTS applications_api_keys;
		DROP TABLE IF EXISTS applications;
	`

	Registry.Register(2, "2_applications_initial_schema", forwards, backwards)
}
