import React from 'react';
import { Tile } from '../interfaces/Tile';

interface LikedTile {
  title: string;
  summary: string;
  imageUrl: string;
}

interface LikedListComponentProps {
  likedTiles: LikedTile[];
}

export const LikedListComponent: React.FC<LikedListComponentProps> = ({ likedTiles }) => {
  return (
    <div>
      <h2>Liked Tiles</h2>
      {likedTiles.map((tile, index) => (
        <div key={index}>
          <img src={tile.imageUrl} alt={tile.title} style={{ width: '100px', height: '50px' }} />
          <h2>{tile.title}</h2>
          <p>{tile.summary}</p>
        </div>
      ))}
    </div>
  );
};
