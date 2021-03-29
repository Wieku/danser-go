#version 330

uniform vec2 cursorPosition;
uniform float radius;
uniform float dim;
uniform float maxDim;

in vec2 osuPosition;

out vec4 color;

void main()
{
    float difference = length(cursorPosition - osuPosition);

    float t = pow(clamp(difference/radius, 0.0, 1.0), 5.0);

    color = vec4(vec3(0.0), mix(t, 1.0, dim)*maxDim);
}