-- Table structure for users
DROP TABLE IF EXISTS users;
CREATE TABLE users (
  id BIGINT PRIMARY KEY NOT NULL,
  name VARCHAR NOT NULL DEFAULT '',
  avatar VARCHAR NOT NULL DEFAULT '',
  background_image VARCHAR NOT NULL DEFAULT '',
  signature VARCHAR NOT NULL DEFAULT '',
  create_time TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_user_name ON users (name);

-- Add comments
COMMENT ON COLUMN users.name IS '用户名';
COMMENT ON COLUMN users.avatar IS '头像地址';
COMMENT ON COLUMN users.background_image IS '背景图地址';
COMMENT ON COLUMN users.signature IS '个性签名';
COMMENT ON COLUMN users.create_time IS '创建时间';
COMMENT ON COLUMN users.update_time IS '更新时间';

-- Table structure for user_logins
DROP TABLE IF EXISTS user_logins;
CREATE TABLE user_logins (
  id BIGINT PRIMARY KEY NOT NULL,
  username VARCHAR NOT NULL,
  password VARCHAR NOT NULL DEFAULT '',
  create_time TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_time TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_login_name ON user_logins (username);

-- Add comments
COMMENT ON COLUMN user_logins.username IS '用户名';
COMMENT ON COLUMN user_logins.password IS '加密密码';
COMMENT ON COLUMN user_logins.create_time IS '创建时间';
COMMENT ON COLUMN user_logins.update_time IS '更新时间';

-- Table structure for videos
DROP TABLE IF EXISTS videos;
CREATE TABLE videos (
  id BIGINT PRIMARY KEY NOT NULL,
  author_id BIGINT NOT NULL DEFAULT 0,
  play_url VARCHAR NOT NULL DEFAULT '',
  cover_url VARCHAR NOT NULL DEFAULT '',
  upload_time TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  title VARCHAR NOT NULL DEFAULT ''
);

CREATE INDEX idx_author_id ON videos (author_id);
CREATE INDEX idx_videos_upload_time ON videos (upload_time);

-- Add comments
COMMENT ON COLUMN videos.author_id IS '作者ID';
COMMENT ON COLUMN videos.play_url IS '视频地址';
COMMENT ON COLUMN videos.cover_url IS '封面地址';
COMMENT ON COLUMN videos.upload_time IS '上传时间';
COMMENT ON COLUMN videos.title IS '标题';

-- Table structure for comments
DROP TABLE IF EXISTS comments;
CREATE TABLE comments (
  id BIGINT PRIMARY KEY NOT NULL,
  video_id BIGINT NOT NULL DEFAULT 0,
  user_id BIGINT NOT NULL DEFAULT 0,
  content VARCHAR NOT NULL DEFAULT '',
  create_time TIMESTAMPTZ(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_comment_video_id ON comments (video_id);
CREATE INDEX idx_user_id ON comments (user_id);

-- Add comments
COMMENT ON COLUMN comments.video_id IS '视频ID';
COMMENT ON COLUMN comments.user_id IS '用户ID';
COMMENT ON COLUMN comments.content IS '评论内容';
COMMENT ON COLUMN comments.create_time IS '创建时间';

-- Table structure for favorites
DROP TABLE IF EXISTS favorites;
CREATE TABLE favorites (
  id BIGINT PRIMARY KEY NOT NULL,
  user_id BIGINT NOT NULL DEFAULT 0,
  video_id BIGINT NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX idx_user_video ON favorites (user_id, video_id);
CREATE INDEX idx_favorite_video_id ON favorites (video_id);

-- Add comments
COMMENT ON COLUMN favorites.user_id IS '用户ID';
COMMENT ON COLUMN favorites.video_id IS '视频ID';

-- Table structure for messages
DROP TABLE IF EXISTS messages;
CREATE TABLE messages (
  id BIGINT PRIMARY KEY NOT NULL,
  from_user_id BIGINT NOT NULL DEFAULT 0,
  to_user_id BIGINT NOT NULL DEFAULT 0,
  convert_id VARCHAR NOT NULL DEFAULT '',
  content VARCHAR NOT NULL DEFAULT '',
  create_time BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX idx_convertId_createTime ON messages (convert_id, create_time);

-- Add comments
COMMENT ON COLUMN messages.from_user_id IS '发送者ID';
COMMENT ON COLUMN messages.to_user_id IS '接收者ID';
COMMENT ON COLUMN messages.convert_id IS '会话ID';
COMMENT ON COLUMN messages.content IS '消息内容';
COMMENT ON COLUMN messages.create_time IS '创建时间';
