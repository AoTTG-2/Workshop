BEGIN;

-- Favorites
CREATE OR REPLACE FUNCTION favorites_count_update()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
UPDATE posts
SET favorites_count = favorites_count + 1
WHERE id = NEW.post_id;
RETURN NEW;

ELSIF TG_OP = 'DELETE' THEN
UPDATE posts
SET favorites_count = favorites_count - 1
WHERE id = OLD.post_id;
RETURN OLD;
END IF;

RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_favorites_count_update
    AFTER INSERT OR DELETE
ON favorites
FOR EACH ROW
EXECUTE PROCEDURE favorites_count_update();

-- Votes
CREATE OR REPLACE FUNCTION votes_count_update()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
UPDATE posts
SET rating = rating + NEW.vote
WHERE id = NEW.post_id;
RETURN NEW;

ELSIF TG_OP = 'DELETE' THEN
UPDATE posts
SET rating = rating - OLD.vote
WHERE id = OLD.post_id;
RETURN OLD;

ELSIF TG_OP = 'UPDATE' THEN
    IF NEW.vote <> OLD.vote THEN
UPDATE posts
SET rating = rating + (NEW.vote - OLD.vote)
WHERE id = NEW.post_id;
END IF;
RETURN NEW;
END IF;

RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_votes_count_update
    AFTER INSERT OR DELETE OR UPDATE
ON votes
FOR EACH ROW
EXECUTE PROCEDURE votes_count_update();

-- Comments
CREATE OR REPLACE FUNCTION comments_count_update()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
UPDATE posts
SET comments_count = comments_count + 1
WHERE id = NEW.post_id;
RETURN NEW;

ELSIF TG_OP = 'DELETE' THEN
UPDATE posts
SET comments_count = comments_count - 1
WHERE id = OLD.post_id;
RETURN OLD;
END IF;

RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_comments_count_update
    AFTER INSERT OR DELETE
ON comments
FOR EACH ROW
EXECUTE PROCEDURE comments_count_update();

COMMIT;
