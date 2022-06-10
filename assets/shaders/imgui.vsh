#version 330

uniform mat4 proj;

in vec2 in_position;
in vec2 in_uv;
in vec4 in_color;

out vec2 uv;
out vec4 color;

void main()
{
    uv = in_uv;
    color = in_color;
    gl_Position = proj * vec4(in_position, 0, 1);
}