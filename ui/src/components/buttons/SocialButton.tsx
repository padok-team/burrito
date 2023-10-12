import React from "react";

import Button from "@/components/buttons/Button";
import GithubIcon from "@/assets/icons/GithubIcon";
import GitlabIcon from "@/assets/icons/GitlabIcon";

interface SocialButtonProps {
  variant: "github" | "gitlab";
  isLoading?: boolean;
}

const SocialButton: React.FC<SocialButtonProps> = ({ variant, isLoading }) => {
  const getContent = () => {
    switch (variant) {
      case "github":
        return (
          <>
            <GithubIcon />
            <span>Login with GitHub</span>
          </>
        );
      case "gitlab":
        return (
          <>
            <GitlabIcon />
            <span>Login with Gitlab</span>
          </>
        );
    }
  };

  return (
    <Button variant="secondary" isLoading={isLoading} className="w-[300px]">
      <div className="flex items-center gap-4 justify-center">
        {getContent()}
      </div>
    </Button>
  );
};

export default SocialButton;
