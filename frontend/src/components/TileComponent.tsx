import React from 'react';
import { Tile } from '../interfaces/Tile';

interface TileProps extends Tile {
  onToggleLike: () => void;
  showLikeButton?: boolean;
}

function formatDate(date: Date | string) {
  if (date instanceof Date) {
    return date.toLocaleString()
  } else {
    return date
  }
}

export const TileComponent: React.FC<TileProps> = ({ name, title, summary, slug, status, publishDate, owner, sharingImage1x1Url, liked, onToggleLike, showLikeButton }) => {
  return (
    <div>
      <img src={sharingImage1x1Url == null? 
        "https://seeeff-prod-static-images.imgix.net/shades_logo.png?w=80&amp;auto=format&amp;dpr=2":  
        sharingImage1x1Url} 
        alt={title} 
        style={{ width: '100px', height: '50px' }} 
      />
      <h2>{title}</h2>
      <p>Slug: {slug}</p>
      <p>Summary: {summary}</p>
      <p>Status: {status}</p>
      <p>Publish Date: {formatDate(publishDate)}</p>
      <p>Owner: {owner}</p>
      { showLikeButton && (
        <button onClick={onToggleLike}>{liked ? '❤️' : '♡'}</button>
      )}
    </div>
  );
};

