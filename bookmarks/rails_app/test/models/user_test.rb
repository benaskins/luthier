require "test_helper"

class UserTest < ActiveSupport::TestCase
  test "has many bookmarks" do
    user = users(:one)
    assert_respond_to user, :bookmarks
    assert_includes user.bookmarks, bookmarks(:one)
  end

  test "dependent destroy removes bookmarks" do
    user = users(:one)
    assert user.bookmarks.count > 0
    user.destroy
    assert_equal 0, Bookmark.where(user_id: user.id).count
  end
end
