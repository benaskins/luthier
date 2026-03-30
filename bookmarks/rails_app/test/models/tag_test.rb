require "test_helper"

class TagTest < ActiveSupport::TestCase
  test "valid tag" do
    tag = Tag.new(name: "programming")
    assert tag.valid?
  end

  test "invalid without name" do
    tag = Tag.new
    assert_not tag.valid?
    assert_includes tag.errors[:name], "can't be blank"
  end

  test "invalid with duplicate name" do
    Tag.create!(name: "unique-tag")
    tag = Tag.new(name: "unique-tag")
    assert_not tag.valid?
    assert tag.errors[:name].any?
  end

  test "case-insensitive uniqueness" do
    Tag.create!(name: "Programming")
    tag = Tag.new(name: "programming")
    assert_not tag.valid?
  end

  test "can have bookmarks" do
    tag = tags(:ruby)
    bookmark = bookmarks(:one)
    tag.bookmarks << bookmark
    assert_includes tag.bookmarks, bookmark
  end
end
