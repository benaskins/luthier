class CreateJoinTableBookmarksTags < ActiveRecord::Migration[8.1]
  def change
    create_join_table :bookmarks, :tags do |t|
      t.index [:bookmark_id, :tag_id], unique: true
      t.index [:tag_id, :bookmark_id]
    end
  end
end
