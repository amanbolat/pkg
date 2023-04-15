package pkgsql

import (
	"fmt"
)

// EndTx ends the Tx and rollbacks it if an err is not nil,
// otherwise the function commits the transaction
func EndTx(tx Tx, err error) error {
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return fmt.Errorf("failed to rollback transaction :%w", err)
		}

		return err
	}

	if cErr := tx.Commit(); cErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
