import React from 'react';

function RefreshButton() {
  const refreshPage = () => {
    window.location.reload();
  };

  return (
    <button onClick={refreshPage}>
      Refresh Page
    </button>
  );
}

export default RefreshButton;