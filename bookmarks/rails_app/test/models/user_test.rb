require "test_helper"

class UserTest < ActiveSupport::TestCase
  test "valid user with email and password" do
    user = User.new(email: "test@example.com", password: "password123", password_confirmation: "password123")
    assert user.valid?
  end

  test "invalid without email" do
    user = User.new(password: "password123")
    assert_not user.valid?
    assert_includes user.errors[:email], "can't be blank"
  end

  test "invalid with duplicate email" do
    User.create!(email: "dupe@example.com", password: "password123")
    user = User.new(email: "dupe@example.com", password: "password456")
    assert_not user.valid?
    assert_includes user.errors[:email], "has already been taken"
  end

  test "invalid with short password" do
    user = User.new(email: "short@example.com", password: "abc")
    assert_not user.valid?
    assert user.errors[:password].any?
  end

  test "authenticates with correct password" do
    user = User.create!(email: "auth@example.com", password: "correctpass")
    assert user.valid_password?("correctpass")
  end

  test "does not authenticate with wrong password" do
    user = User.create!(email: "auth2@example.com", password: "correctpass")
    assert_not user.valid_password?("wrongpass")
  end
end
