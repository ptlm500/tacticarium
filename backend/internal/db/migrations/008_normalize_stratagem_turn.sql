-- +goose Up
-- Normalize curly right-single-quotes (U+2019) to straight apostrophes in the
-- stratagems.turn column so the frontend's string comparisons against
-- "Your turn" / "Opponent's turn" / "Either player's turn" match.
UPDATE stratagems
SET turn = REPLACE(turn, '’', '''')
WHERE turn LIKE '%’%';

-- +goose Down
-- No-op: reverting to curly apostrophes would re-break turn filtering.
SELECT 1;
