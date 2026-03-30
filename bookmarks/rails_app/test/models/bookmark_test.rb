require "test_helper"

class BookmarkTest < ActiveSupport::TestCase
  setup do
    @user = users(:one)
  end

  test "valid bookmark" do
    bookmark = Bookmark.new(url: "https://example.com", title: "Example", user: @user)
    assert bookmark.valid?
  end

  test "invalid without title" do
    bookmark = Bookmark.new(url: "https://example.com", user: @user)
    assert_not bookmark.valid?
    assert_includes bookmark.errors[:title], "can't be blank"
  end

  test "invalid without url" do
    bookmark = Bookmark.new(title: "Example", user: @user)
    assert_not bookmark.valid?
    assert_includes bookmark.errors[:url], "can't be blank"
  end

  test "invalid with non-http url" do
    bookmark = Bookmark.new(url: "ftp://example.com", title: "Example", user: @user)
    assert_not bookmark.valid?
    assert bookmark.errors[:url].any?
  end

  test "invalid with plain string url" do
    bookmark = Bookmark.new(url: "not-a-url", title: "Example", user: @user)
    assert_not bookmark.valid?
    assert bookmark.errors[:url].any?
  end

  test "valid with http url" do
    bookmark = Bookmark.new(url: "http://example.com", title: "Example", user: @user)
    assert bookmark.valid?
  end

  test "valid with https url" do
    bookmark = Bookmark.new(url: "https://example.com/path?q=1", title: "Example", user: @user)
    assert bookmark.valid?
  end

  test "belongs to user" do
    bookmark = bookmarks(:one)
    assert_equal users(:one), bookmark.user
  end

  test "can have tags" do
    bookmark = bookmarks(:one)
    tag = tags(:ruby)
    bookmark.tags << tag
    assert_includes bookmark.tags, tag
  end

  test "many-to-many with tags" do
    bookmark = bookmarks(:one)
    bookmark.tags << tags(:ruby)
    bookmark.tags << tags(:rails)
    assert_equal 2, bookmark.tags.count
  end
end
