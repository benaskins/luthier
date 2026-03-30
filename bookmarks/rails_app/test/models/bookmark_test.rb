require "test_helper"

class BookmarkTest < ActiveSupport::TestCase
  test "valid bookmark" do
    bookmark = Bookmark.new(
      url: "https://example.com",
      title: "Example",
      user: users(:one)
    )
    assert bookmark.valid?
  end

  test "requires url" do
    bookmark = Bookmark.new(title: "Example", user: users(:one))
    assert_not bookmark.valid?
    assert_includes bookmark.errors[:url], "can't be blank"
  end

  test "requires title" do
    bookmark = Bookmark.new(url: "https://example.com", user: users(:one))
    assert_not bookmark.valid?
    assert_includes bookmark.errors[:title], "can't be blank"
  end

  test "requires valid url format" do
    bookmark = Bookmark.new(
      url: "not-a-url",
      title: "Bad URL",
      user: users(:one)
    )
    assert_not bookmark.valid?
    assert_includes bookmark.errors[:url], "must be a valid HTTP or HTTPS URL"
  end

  test "accepts http urls" do
    bookmark = Bookmark.new(
      url: "http://example.com",
      title: "HTTP",
      user: users(:one)
    )
    assert bookmark.valid?
  end

  test "accepts https urls" do
    bookmark = Bookmark.new(
      url: "https://example.com",
      title: "HTTPS",
      user: users(:one)
    )
    assert bookmark.valid?
  end

  test "belongs to user" do
    bookmark = bookmarks(:one)
    assert_equal users(:one), bookmark.user
  end

  test "has and belongs to many tags" do
    bookmark = bookmarks(:one)
    tag = tags(:ruby)
    bookmark.tags << tag
    assert_includes bookmark.reload.tags, tag
  end

  test "for_user scope returns only user bookmarks" do
    user_one_bookmarks = Bookmark.for_user(users(:one))
    assert_includes user_one_bookmarks, bookmarks(:one)
    assert_includes user_one_bookmarks, bookmarks(:two)
    assert_not_includes user_one_bookmarks, bookmarks(:other_user_bookmark)
  end

  test "newest_first scope orders by created_at desc" do
    bookmarks = Bookmark.for_user(users(:one)).newest_first
    assert bookmarks.first.created_at >= bookmarks.last.created_at
  end

  test "oldest_first scope orders by created_at asc" do
    bookmarks = Bookmark.for_user(users(:one)).oldest_first
    assert bookmarks.first.created_at <= bookmarks.last.created_at
  end

  test "destroying user destroys bookmarks" do
    user = users(:one)
    bookmark_ids = user.bookmarks.pluck(:id)
    assert bookmark_ids.any?
    user.destroy
    bookmark_ids.each do |id|
      assert_nil Bookmark.find_by(id: id)
    end
  end
end
