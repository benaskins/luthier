class Bookmark < ApplicationRecord
  belongs_to :user
  has_and_belongs_to_many :tags

  validates :url, presence: true, format: {
    with: /\Ahttps?:\/\/.+/,
    message: "must be a valid HTTP or HTTPS URL"
  }
  validates :title, presence: true

  scope :for_user, ->(user) { where(user: user) }
  scope :newest_first, -> { order(created_at: :desc) }
  scope :oldest_first, -> { order(created_at: :asc) }
  scope :recently_updated, -> { order(updated_at: :desc) }
end
