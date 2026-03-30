require "test_helper"

class TagTest < ActiveSupport::TestCase
  test "valid tag" do
    tag = Tag.new(name: "javascript")
    assert tag.valid?
  end

  test "requires name" do
    tag = Tag.new(name: nil)
    assert_not tag.valid?
    assert_includes tag.errors[:name], "can't be blank"
  end

  test "requires unique name case insensitive" do
    Tag.create!(name: "Unique Tag")
    duplicate = Tag.new(name: "unique tag")
    assert_not duplicate.valid?
    assert_includes duplicate.errors[:name], "has already been taken"
  end

  test "has and belongs to many bookmarks" do
    tag = tags(:ruby)
    bookmark = bookmarks(:one)
    tag.bookmarks << bookmark
    assert_includes tag.reload.bookmarks, bookmark
  end
end
