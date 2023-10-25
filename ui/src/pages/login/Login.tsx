import React, { useContext } from "react";
import { useNavigate } from "react-router-dom";

import { ThemeContext } from "@/contexts/ThemeContext";

import Input from "@/components/inputs/Input";
import Button from "@/components/buttons/Button";
import SocialButton from "@/components/buttons/SocialButton";

import Burrito from "@/assets/illustrations/Burrito";
import EyeSlashIcon from "@/assets/icons/EyeSlashIcon";
// import CoverLight from "@/assets/covers/cover-light.png";
// import CoverDark from "@/assets/covers/cover-dark.png";

const Login: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const navigate = useNavigate();
  return (
    <div className="flex h-screen">
      <div
        className={`
          inline-block
          justify-center
          overflow-y-scroll
          w-[500px]
          min-w-[500px]
          p-6
          ${theme === "light" ? "bg-nuances-white" : "bg-nuances-black"}
        `}
      >
        <div className="flex items-center justify-center">
          <div className="flex flex-col items-start justify-center min-w-[300px] gap-10">
            <div className="flex flex-col items-start gap-2">
              <Burrito height={64} width={64} />
              <span
                className={`
                  text-3xl
                  font-extrabold
                  ${
                    theme === "light"
                      ? "text-nuances-black"
                      : "text-nuances-white"
                  }
                `}
              >
                Welcome to Burrito
              </span>
            </div>
            <div className="flex flex-col items-center justify-center gap-8">
              <Input
                className="min-w-full"
                variant={theme}
                placeholder="Your email"
                label="Email"
              />
              <Input
                className="min-w-full"
                variant={theme}
                placeholder="Your password"
                label="Password"
                rightIcon={<EyeSlashIcon />}
              />
              <Button
                className="min-w-full"
                variant={theme === "light" ? "primary" : "secondary"}
                onClick={() => navigate("/")}
              >
                Login
              </Button>
              <div
                className={`
                  flex
                  flex-row
                  items-center
                  min-w-full
                  gap-4
                  ${
                    theme === "light"
                      ? "text-nuances-black border-nuances-black"
                      : "text-nuances-white border-nuances-white"
                  }
                `}
              >
                <hr className="w-full" />
                <span>OR</span>
                <hr className="w-full" />
              </div>
              <div className="flex flex-col items-center justify-center min-w-full gap-4">
                <SocialButton
                  className="w-full"
                  variant="github"
                  onClick={() => navigate("/")}
                />
                <SocialButton
                  className="w-full"
                  variant="gitlab"
                  onClick={() => navigate("/")}
                />
              </div>
            </div>
            <div
              className={`
                flex
                flex-row
                items-center
                justify-center
                min-w-full
                gap-1
                p-6
                rounded-lg
                ${
                  theme === "light"
                    ? "bg-primary-400 text-nuances-black"
                    : "bg-nuances-400 text-nuances-50"
                }
              `}
            >
              <span className="text-base font-normal">
                Don't have an account ?
              </span>
              <span className="text-base font-semibold">Sign up</span>
            </div>
          </div>
        </div>
      </div>
      <div
        className={`
          relative
          flex
          flex-col
          w-[calc(100%_-_500px)]
          pt-20
          px-16
          gap-6
          bg-background-dark
        `}
      >
        <div
          className={`
            flex
            flex-col
            ${theme === "light" ? "text-nuances-black" : "text-nuances-white"}
          `}
        >
          <span className="text-5xl font-extrabold">Burrito is a TACoS</span>
          <span className="text-base font-medium">
            Monitor the status of your layers and their impacts on your project.
          </span>
        </div>
        {/* <img
          className="absolute right-0 bottom-0 -rotate-12 translate-y-20 translate-x-28 shadow-light rounded-lg h-[600px] object-cover"
          src={CoverLight}
        /> */}
      </div>
    </div>
  );
};

export default Login;
