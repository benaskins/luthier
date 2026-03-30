class Bookmark < ApplicationRecord
  belongs_to :user
  has_and_belongs_to_many :tags

  validates :title, presence: true
  validates :url, presence: true, format: {
    with: /\Ahttps?:\/\/.+/,
    message: "must be a valid HTTP or HTTPS URL"
  }
end
