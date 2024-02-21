import React, { useState, useEffect } from 'react';
import { Tile } from './interfaces/Tile'
import { SearchComponent } from './components/SearchComponent';
import { TileComponent } from './components/TileComponent';
import RefreshButton from './components/RefreshButton';

export const App: React.FC = () => {
  const [tiles, setTiles] = useState<Tile[]>([]);
  const [likedTiles, setLikedTiles] = useState<Tile[]>([]);
  const [titlesOfLiked, setLikedSet] = useState<Set<string>>(new Set());
  const titleSet = new Set(titlesOfLiked);

  const handleSearch = async (searchTerm: string) => {
    console.log("Handle search called with term:", searchTerm);
    try {
        const response = await fetch(`/search?q=${searchTerm}`); 
        if (!response.ok) {
            throw new Error('Network response was not ok ' + response.statusText);
        }
        const data: any = await response.json();
        console.log("Response data:", data); 
        
        setTiles(data.data.result.map((tile: any) => ({
            name: tile.name || 'Unknown', 
            title: tile.title,
            summary: tile.summary,
            slug: tile.slug.current,
            status: tile.status,
            publishDate: new Date(tile.publishDate),
            owner: tile.owner,
            sharingImage1x1Url: tile.sharingImage1x1Url, 
            liked: titleSet.has(tile.title)? true: false,
        })));
    } catch (error) {
      console.error('Error fetching data:', error);
    }
  };
  
  function updateSearchLikes() {
    titleSet.clear();
    for (const tile of likedTiles) {
      titleSet.add(tile.title);
    }
    console.log(`likedSet: ${[...titleSet]}`)
    setLikedSet(titleSet);
    let newArr = [...tiles];
    for (const tile of newArr) {
      if (titleSet.has(tile.title)) {
        tile.liked = true;
      } else {
        tile.liked = false;
      }
    }
    setTiles(newArr);
  }

  useEffect(() => {
    updateSearchLikes();
  }, [likedTiles]);

  const handleToggleLike = async (tile: Tile | null) => {
    try {
        const endpoint = tile == null || tile.liked? 'unlike': 'like';
        const body = tile == null? '{}': JSON.stringify(tile);
        console.log(`Hitting endpoint ${endpoint}`)
        const response = await fetch(`/${endpoint}`, { 
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: body,
        });
        if (!response.ok) {
            throw new Error('Failed to fetch data');
        }

        const data = await response.json();
        console.log(`response: ${response}, data: ${JSON.stringify(data)}`)
        // Assuming the response contains a list of JSON objects representing tiles
        setLikedTiles(data);



    } catch (error) {
        console.error('Error:', error);
    }
  };

  const updateLikes = async() => {
    const response = await fetch(`/get_liked`);
    const data = await response.json();
    setLikedTiles(data);
  }

  useEffect(() => {
    updateLikes();
  }, []);

  return (
    <div style={{
      display: 'flex',
      flexDirection: 'row',
      justifyContent: 'space-around',
      height: '100vh',
      padding: '20px'
    }}>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'flex-start',
        width: '50%',
        paddingTop: '20px'
      }}>
        <SearchComponent onSearch={handleSearch} />
        {tiles.map((tile, index) => (
          <TileComponent
            key={index}
            name={tile.name}
            summary={tile.summary}
            status={tile.status}
            publishDate={tile.publishDate}
            owner={tile.owner}
            title={tile.title}
            slug={tile.slug}
            sharingImage1x1Url={tile.sharingImage1x1Url}
            liked={tile.liked}
            onToggleLike={() => handleToggleLike(tile)}
            showLikeButton={true}
          />
        ))}
      </div>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'flex-start',
        width: '50%',
        paddingTop: '20px',
        borderLeft: '1px solid #000'
      }}>
        <h2>Liked Tiles</h2>
        <RefreshButton></RefreshButton>
        {likedTiles.map((tile, index) => (
          <TileComponent
            key={index}
            name={tile.name}
            summary={tile.summary}
            status={tile.status}
            publishDate={tile.publishDate}
            owner={tile.owner}
            title={tile.title}
            slug={tile.slug}
            sharingImage1x1Url={tile.sharingImage1x1Url}
            liked={tile.liked}
            onToggleLike={() => handleToggleLike(tile)}
            showLikeButton={false}
          />
        ))}
      </div>
    </div>
  );
};

export default App;
